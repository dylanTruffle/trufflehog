#!/usr/bin/env python3
"""
TruffleHog Network Verification Analysis
Analyzes timeouts, unreachable hosts, and connection patterns
"""

import json
import re
from collections import defaultdict, Counter
from urllib.parse import urlparse
import matplotlib.pyplot as plt
import seaborn as sns
import pandas as pd

def load_scan_data(filename):
    """Load and parse scan data"""
    secrets = []
    logs = []
    
    try:
        with open(filename, 'r') as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    data = json.loads(line)
                    if 'DetectorType' in data:
                        secrets.append(data)
                    elif 'msg' in data:
                        logs.append(data)
                except json.JSONDecodeError:
                    continue
    except FileNotFoundError:
        print(f"Warning: {filename} not found")
        return [], []
    
    return secrets, logs

def extract_host_from_error(error_msg):
    """Extract host/IP from verification error messages"""
    if not error_msg:
        return None
    
    # Pattern for "dial tcp HOST:PORT: error"
    tcp_pattern = r'dial tcp ([^:]+):\d+:'
    match = re.search(tcp_pattern, error_msg)
    if match:
        return match.group(1)
    
    # Pattern for HTTP URLs
    url_pattern = r'https?://([^/]+)'
    match = re.search(url_pattern, error_msg)
    if match:
        return match.group(1)
    
    return None

def get_detector_host_mapping():
    """Map detector types to their typical verification hosts"""
    # Based on common detector patterns and the errors we see
    detector_hosts = {
        'JDBC': ['database hosts', 'mysql hosts'],
        'Gitlab': ['gitlab.com', 'gitlab instances'],
        'GitHub': ['github.com', 'api.github.com'],
        'Slack': ['slack.com', 'hooks.slack.com'],
        'CircleCI': ['circleci.com'],
        'BuildKite': ['api.buildkite.com'],
        'CloudflareApiToken': ['api.cloudflare.com'],
        'URI': ['various extracted URLs'],
        'PrivateKey': ['SSH hosts', 'TLS hosts']
    }
    return detector_hosts

def analyze_verification_errors(secrets):
    """Analyze verification errors and timeouts"""
    error_analysis = {
        'timeout_errors': [],
        'connection_refused': [],
        'unreachable_hosts': [],
        'unknown_hosts': [],
        'other_errors': [],
        'successful_verifications': []
    }
    
    host_error_counts = defaultdict(int)
    detector_error_patterns = defaultdict(list)
    
    for secret in secrets:
        verified = secret.get('Verified', False)
        verification_error = secret.get('VerificationError', '')
        detector_name = secret.get('DetectorName', 'Unknown')
        
        if verified:
            error_analysis['successful_verifications'].append(secret)
        elif verification_error:
            # Extract host from error
            host = extract_host_from_error(verification_error)
            if host:
                host_error_counts[host] += 1
            
            detector_error_patterns[detector_name].append(verification_error)
            
            # Categorize errors
            error_lower = verification_error.lower()
            if 'timeout' in error_lower or 'i/o timeout' in error_lower:
                error_analysis['timeout_errors'].append({
                    'secret': secret,
                    'host': host,
                    'detector': detector_name,
                    'error': verification_error
                })
            elif 'connection refused' in error_lower:
                error_analysis['connection_refused'].append({
                    'secret': secret,
                    'host': host,
                    'detector': detector_name,
                    'error': verification_error
                })
            elif 'no such host' in error_lower or 'host not found' in error_lower:
                error_analysis['unreachable_hosts'].append({
                    'secret': secret,
                    'host': host,
                    'detector': detector_name,
                    'error': verification_error
                })
            else:
                error_analysis['other_errors'].append({
                    'secret': secret,
                    'host': host,
                    'detector': detector_name,
                    'error': verification_error
                })
    
    return error_analysis, host_error_counts, detector_error_patterns

def estimate_timeout_waste(error_analysis, timeout_duration=5.0):
    """Estimate time wasted on timeouts"""
    timeout_count = len(error_analysis['timeout_errors'])
    connection_refused_count = len(error_analysis['connection_refused'])
    unreachable_count = len(error_analysis['unreachable_hosts'])
    
    # Timeouts take full timeout duration
    timeout_waste = timeout_count * timeout_duration
    
    # Connection refused is usually fast (~0.1s)
    connection_refused_waste = connection_refused_count * 0.1
    
    # Unreachable hosts usually timeout on DNS/connect (~2s average)
    unreachable_waste = unreachable_count * 2.0
    
    total_waste = timeout_waste + connection_refused_waste + unreachable_waste
    
    return {
        'timeout_waste_seconds': timeout_waste,
        'connection_refused_waste_seconds': connection_refused_waste,
        'unreachable_waste_seconds': unreachable_waste,
        'total_waste_seconds': total_waste,
        'timeout_count': timeout_count,
        'connection_refused_count': connection_refused_count,
        'unreachable_count': unreachable_count
    }

def analyze_host_blacklist_savings(host_error_counts, error_analysis):
    """Analyze potential savings from host blacklisting"""
    failed_hosts = set()
    
    # Collect all failed hosts
    for error_type in ['timeout_errors', 'connection_refused', 'unreachable_hosts']:
        for error in error_analysis[error_type]:
            if error['host']:
                failed_hosts.add(error['host'])
    
    # Calculate savings if we remembered these hosts
    total_failed_attempts = sum(host_error_counts.values())
    
    # Potential savings: after first failure, skip subsequent attempts
    potential_savings = 0
    for host, count in host_error_counts.items():
        if count > 1:
            # Save time on all attempts after the first
            potential_savings += (count - 1) * 5.0  # 5 second timeout per attempt
    
    return {
        'failed_hosts': failed_hosts,
        'total_failed_attempts': total_failed_attempts,
        'repeat_failures': sum(max(0, count - 1) for count in host_error_counts.values()),
        'potential_savings_seconds': potential_savings,
        'unique_failed_hosts': len(failed_hosts),
        'host_failure_distribution': dict(host_error_counts)
    }

def analyze_connection_batching_potential(secrets):
    """Analyze potential for connection batching by detector type and host"""
    detector_patterns = defaultdict(list)
    host_detector_map = defaultdict(lambda: defaultdict(list))
    
    for secret in secrets:
        detector_name = secret.get('DetectorName', 'Unknown')
        raw_secret = secret.get('Raw', '')
        
        detector_patterns[detector_name].append(secret)
        
        # Try to extract host patterns from the secret itself
        host = None
        if detector_name == 'JDBC':
            # JDBC URLs contain host info
            jdbc_match = re.search(r'jdbc:[^:]+://([^:/]+)', raw_secret)
            if jdbc_match:
                host = jdbc_match.group(1)
        elif detector_name == 'URI':
            # General URI pattern
            try:
                parsed = urlparse(raw_secret)
                if parsed.netloc:
                    host = parsed.netloc
            except:
                pass
        elif detector_name in ['Gitlab', 'GitHub']:
            # These typically hit their respective APIs
            host = f"{detector_name.lower()}.com"
        
        if host:
            host_detector_map[host][detector_name].append(secret)
    
    # Calculate batching potential
    batching_analysis = {}
    total_potential_savings = 0
    
    for host, detectors in host_detector_map.items():
        for detector, secrets_list in detectors.items():
            if len(secrets_list) > 1:
                # Each connection setup/teardown costs ~0.1-0.5s
                connection_overhead = 0.3  # seconds per connection
                current_connections = len(secrets_list)
                optimized_connections = 1  # batch all together
                
                savings = (current_connections - optimized_connections) * connection_overhead
                total_potential_savings += savings
                
                batching_analysis[f"{host}:{detector}"] = {
                    'secrets_count': len(secrets_list),
                    'current_connections': current_connections,
                    'potential_savings_seconds': savings
                }
    
    return {
        'batching_opportunities': batching_analysis,
        'total_connection_savings_seconds': total_potential_savings,
        'unique_host_detector_pairs': len(batching_analysis),
        'detector_distribution': {k: len(v) for k, v in detector_patterns.items()}
    }

def create_network_analysis_visualizations(error_analysis, timeout_waste, blacklist_savings, batching_analysis):
    """Create visualizations for network analysis"""
    
    # Create a comprehensive figure
    fig, ((ax1, ax2), (ax3, ax4)) = plt.subplots(2, 2, figsize=(16, 12))
    
    # 1. Error Type Distribution
    error_counts = {
        'Timeouts': len(error_analysis['timeout_errors']),
        'Connection Refused': len(error_analysis['connection_refused']),
        'Unreachable Hosts': len(error_analysis['unreachable_hosts']),
        'Other Errors': len(error_analysis['other_errors']),
        'Successful': len(error_analysis['successful_verifications'])
    }
    
    colors = ['#FF6B6B', '#FFA500', '#FFD700', '#FF69B4', '#90EE90']
    wedges, texts, autotexts = ax1.pie(error_counts.values(), labels=error_counts.keys(), 
                                       autopct='%1.1f%%', colors=colors, startangle=90)
    ax1.set_title('Verification Results Distribution', fontsize=14, weight='bold')
    
    # Highlight timeouts
    wedges[0].set_edgecolor('red')
    wedges[0].set_linewidth(3)
    
    # 2. Time Waste Breakdown
    waste_categories = ['Timeout Waste', 'Connection Refused', 'Unreachable Hosts']
    waste_values = [
        timeout_waste['timeout_waste_seconds'],
        timeout_waste['connection_refused_waste_seconds'],
        timeout_waste['unreachable_waste_seconds']
    ]
    
    bars = ax2.bar(waste_categories, waste_values, color=['#FF6B6B', '#FFA500', '#FFD700'])
    ax2.set_title('Time Wasted by Error Type (seconds)', fontsize=14, weight='bold')
    ax2.set_ylabel('Seconds Wasted')
    
    # Add value labels on bars
    for bar, value in zip(bars, waste_values):
        height = bar.get_height()
        ax2.text(bar.get_x() + bar.get_width()/2., height + 0.1,
                f'{value:.1f}s', ha='center', va='bottom', fontsize=10, weight='bold')
    
    # 3. Host Blacklist Savings Potential
    savings_categories = ['Current Waste', 'Potential Savings', 'Remaining Waste']
    current_waste = timeout_waste['total_waste_seconds']
    potential_savings = blacklist_savings['potential_savings_seconds']
    remaining_waste = current_waste - potential_savings
    
    savings_values = [current_waste, potential_savings, remaining_waste]
    colors_savings = ['#FF6B6B', '#90EE90', '#FFB347']
    
    bars = ax3.bar(savings_categories, savings_values, color=colors_savings)
    ax3.set_title('Host Blacklisting Savings Potential', fontsize=14, weight='bold')
    ax3.set_ylabel('Time (seconds)')
    
    # Add improvement annotation
    if potential_savings > 0:
        improvement_pct = (potential_savings / current_waste) * 100
        ax3.annotate(f'{improvement_pct:.1f}% improvement\nwith host blacklisting', 
                    xy=(1, potential_savings), xytext=(1, potential_savings + current_waste * 0.2),
                    arrowprops=dict(arrowstyle='->', color='green', lw=2),
                    fontsize=11, ha='center', color='green', weight='bold')
    
    # 4. Connection Batching Potential
    batching_categories = ['Current\nConnection Overhead', 'Optimized\nConnection Overhead', 'Savings']
    current_overhead = batching_analysis['total_connection_savings_seconds'] + 1  # baseline
    optimized_overhead = 1  # minimal connections
    savings = batching_analysis['total_connection_savings_seconds']
    
    batching_values = [current_overhead, optimized_overhead, savings]
    colors_batch = ['#FF6B6B', '#90EE90', '#87CEEB']
    
    bars = ax4.bar(batching_categories, batching_values, color=colors_batch)
    ax4.set_title('Connection Batching Savings Potential', fontsize=14, weight='bold')
    ax4.set_ylabel('Time (seconds)')
    
    plt.tight_layout()
    plt.savefig('benchmarks/network_analysis.png', dpi=300, bbox_inches='tight')
    plt.close()

def create_detailed_host_analysis(host_error_counts, error_analysis):
    """Create detailed analysis of problematic hosts"""
    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(16, 8))
    
    # 1. Top Failed Hosts
    if host_error_counts:
        top_hosts = dict(Counter(host_error_counts).most_common(10))
        
        bars = ax1.bar(range(len(top_hosts)), list(top_hosts.values()), 
                      color='#FF6B6B', alpha=0.7)
        ax1.set_title('Top 10 Most Failed Hosts', fontsize=14, weight='bold')
        ax1.set_ylabel('Number of Failed Attempts')
        ax1.set_xlabel('Hosts')
        ax1.set_xticks(range(len(top_hosts)))
        ax1.set_xticklabels([host[:20] + '...' if len(host) > 20 else host 
                            for host in top_hosts.keys()], rotation=45, ha='right')
        
        # Add value labels
        for i, (bar, value) in enumerate(zip(bars, top_hosts.values())):
            height = bar.get_height()
            ax1.text(bar.get_x() + bar.get_width()/2., height + 0.1,
                    f'{value}', ha='center', va='bottom', fontsize=9)
    
    # 2. Error Pattern Timeline (if we had timestamps, simulate distribution)
    error_types = ['Timeouts', 'Connection Refused', 'Unreachable', 'Other']
    error_counts_by_type = [
        len(error_analysis['timeout_errors']),
        len(error_analysis['connection_refused']),
        len(error_analysis['unreachable_hosts']),
        len(error_analysis['other_errors'])
    ]
    
    bars = ax2.bar(error_types, error_counts_by_type, 
                  color=['#FF6B6B', '#FFA500', '#FFD700', '#FF69B4'])
    ax2.set_title('Network Error Types Distribution', fontsize=14, weight='bold')
    ax2.set_ylabel('Number of Errors')
    ax2.tick_params(axis='x', rotation=45)
    
    plt.tight_layout()
    plt.savefig('benchmarks/host_analysis.png', dpi=300, bbox_inches='tight')
    plt.close()

def generate_network_recommendations(timeout_waste, blacklist_savings, batching_analysis, total_secrets):
    """Generate specific network optimization recommendations"""
    
    total_waste = timeout_waste['total_waste_seconds']
    blacklist_savings_potential = blacklist_savings['potential_savings_seconds']
    batching_savings_potential = batching_analysis['total_connection_savings_seconds']
    
    total_potential_savings = blacklist_savings_potential + batching_savings_potential
    
    recommendations = f"""
# Network Verification Optimization Recommendations

## 🔍 Analysis Summary
- **Total Secrets Analyzed**: {total_secrets}
- **Total Network Time Wasted**: {total_waste:.1f} seconds
- **Potential Savings**: {total_potential_savings:.1f} seconds ({(total_potential_savings/total_waste)*100:.1f}% improvement)

## 🚨 Critical Network Issues

### 1. Timeout Waste Analysis
- **Timeout Events**: {timeout_waste['timeout_count']} timeouts
- **Time Wasted on Timeouts**: {timeout_waste['timeout_waste_seconds']:.1f} seconds
- **Connection Refused**: {timeout_waste['connection_refused_count']} events ({timeout_waste['connection_refused_waste_seconds']:.1f}s wasted)
- **Unreachable Hosts**: {timeout_waste['unreachable_count']} events ({timeout_waste['unreachable_waste_seconds']:.1f}s wasted)

### 2. Host Blacklisting Potential
- **Unique Failed Hosts**: {blacklist_savings['unique_failed_hosts']} hosts
- **Repeat Failures**: {blacklist_savings['repeat_failures']} redundant attempts
- **Potential Savings**: {blacklist_savings_potential:.1f} seconds ({(blacklist_savings_potential/total_waste)*100:.1f}% improvement)

### 3. Connection Batching Potential
- **Batching Opportunities**: {batching_analysis['unique_host_detector_pairs']} host-detector pairs
- **Connection Overhead Savings**: {batching_savings_potential:.1f} seconds
- **Multiple Connections to Same Host**: Detected across {len(batching_analysis['batching_opportunities'])} cases

## 💡 Implementation Recommendations

### Priority 1: Host Blacklisting (Immediate - 1 week)
```go
type HostBlacklist struct {{
    failedHosts map[string]time.Time
    mu          sync.RWMutex
    blacklistDuration time.Duration
}}

func (hb *HostBlacklist) IsBlacklisted(host string) bool {{
    hb.mu.RLock()
    defer hb.mu.RUnlock()
    
    if failTime, exists := hb.failedHosts[host]; exists {{
        if time.Since(failTime) < hb.blacklistDuration {{
            return true // Skip this host
        }}
        // Remove expired blacklist entry
        delete(hb.failedHosts, host)
    }}
    return false
}}
```

**Expected Impact**: {(blacklist_savings_potential/total_waste)*100:.1f}% reduction in verification time

### Priority 2: Connection Batching (Medium - 2-3 weeks)
```go
type HostBatcher struct {{
    pendingSecrets map[string][]SecretToVerify
    batchTimeout   time.Duration
    maxBatchSize   int
}}

func (hb *HostBatcher) BatchVerify(host string, secrets []Secret) {{
    // Batch multiple secrets for the same host
    // Reuse HTTP connections and database connections
    // Process all secrets in one verification call
}}
```

**Expected Impact**: {(batching_savings_potential/total_waste)*100:.1f}% reduction in connection overhead

### Priority 3: Smart Timeout Configuration (Quick win - 2-3 days)
```go
type AdaptiveTimeout struct {{
    hostTimeouts map[string]time.Duration
    defaultTimeout time.Duration
    maxTimeout     time.Duration
}}

func (at *AdaptiveTimeout) GetTimeout(host string) time.Duration {{
    // Start with short timeout for new hosts
    // Increase timeout for hosts that respond slowly
    // Use very short timeout for previously failed hosts
}}
```

**Expected Impact**: 10-15% reduction in timeout waste

## 📊 Performance Projections

| Optimization | Current Time | Optimized Time | Improvement |
|--------------|--------------|----------------|-------------|
| Host Blacklisting | {total_waste:.1f}s | {total_waste - blacklist_savings_potential:.1f}s | {(blacklist_savings_potential/total_waste)*100:.1f}% |
| Connection Batching | {total_waste:.1f}s | {total_waste - batching_savings_potential:.1f}s | {(batching_savings_potential/total_waste)*100:.1f}% |
| Combined Optimizations | {total_waste:.1f}s | {total_waste - total_potential_savings:.1f}s | {(total_potential_savings/total_waste)*100:.1f}% |

## 🎯 Success Metrics
- **Host Blacklist Hit Rate**: Target >60% for repeat failures
- **Connection Reuse Rate**: Target >80% for same-host requests  
- **Average Verification Time**: Reduce from current to <2s per secret
- **Timeout Rate**: Reduce from current to <5% of all verifications

## 🔧 Monitoring Requirements
1. **Host Failure Tracking**: Track which hosts fail consistently
2. **Connection Pool Metrics**: Monitor connection reuse rates
3. **Timeout Distribution**: Track timeout patterns by host and detector
4. **Batch Efficiency**: Measure batching success rates

---
*Analysis based on Figma scan data with {total_secrets} secrets*
"""
    
    return recommendations

def main():
    """Main analysis function"""
    print("🔍 Starting TruffleHog Network Verification Analysis...")
    
    # Load data
    secrets, logs = load_scan_data('figma_comprehensive_scan.json')
    
    if not secrets:
        print("❌ No scan data found. Please run the Figma scan first.")
        return
    
    print(f"📊 Analyzing {len(secrets)} secrets for network patterns...")
    
    # Perform analysis
    error_analysis, host_error_counts, detector_error_patterns = analyze_verification_errors(secrets)
    timeout_waste = estimate_timeout_waste(error_analysis)
    blacklist_savings = analyze_host_blacklist_savings(host_error_counts, error_analysis)
    batching_analysis = analyze_connection_batching_potential(secrets)
    
    # Create visualizations
    print("📈 Generating network analysis visualizations...")
    create_network_analysis_visualizations(error_analysis, timeout_waste, blacklist_savings, batching_analysis)
    create_detailed_host_analysis(host_error_counts, error_analysis)
    
    # Generate recommendations
    recommendations = generate_network_recommendations(
        timeout_waste, blacklist_savings, batching_analysis, len(secrets)
    )
    
    # Save recommendations
    with open('benchmarks/NETWORK_OPTIMIZATION_ANALYSIS.md', 'w') as f:
        f.write(recommendations)
    
    # Print summary
    print("\n🎯 NETWORK ANALYSIS SUMMARY")
    print("=" * 50)
    print(f"Total Time Wasted: {timeout_waste['total_waste_seconds']:.1f} seconds")
    print(f"Host Blacklisting Savings: {blacklist_savings['potential_savings_seconds']:.1f}s ({(blacklist_savings['potential_savings_seconds']/timeout_waste['total_waste_seconds']*100):.1f}%)")
    print(f"Connection Batching Savings: {batching_analysis['total_connection_savings_seconds']:.1f}s")
    print(f"Total Potential Improvement: {(blacklist_savings['potential_savings_seconds'] + batching_analysis['total_connection_savings_seconds'])/timeout_waste['total_waste_seconds']*100:.1f}%")
    print(f"\n📁 Generated files:")
    print(f"  - benchmarks/network_analysis.png")
    print(f"  - benchmarks/host_analysis.png") 
    print(f"  - benchmarks/NETWORK_OPTIMIZATION_ANALYSIS.md")

if __name__ == "__main__":
    main()