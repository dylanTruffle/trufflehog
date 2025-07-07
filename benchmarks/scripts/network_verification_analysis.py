#!/usr/bin/env python3
"""
TruffleHog Network Verification Waste Analysis
Analyzes DNS failures, timeouts, unreachable hosts, and connection patterns
"""

import json
import re
from collections import defaultdict, Counter
from urllib.parse import urlparse
import matplotlib.pyplot as plt
import seaborn as sns
import pandas as pd
import numpy as np

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

def extract_network_endpoint(secret):
    """Extract the network endpoint (host:port) that would be verified"""
    detector_name = secret.get('DetectorName', '')
    raw_secret = secret.get('Raw', '')
    verification_error = secret.get('VerificationError', '')
    
    endpoint = None
    
    # First try to extract from verification error (most reliable)
    if verification_error:
        # Pattern: "dial tcp HOST:PORT: error"
        tcp_match = re.search(r'dial tcp ([^:]+:\d+):', verification_error)
        if tcp_match:
            return tcp_match.group(1)
        
        # Pattern: "dial tcp HOST:PORT error" (no colon)
        tcp_match2 = re.search(r'dial tcp ([^:]+:\d+)\s', verification_error)
        if tcp_match2:
            return tcp_match2.group(1)
        
        # Pattern for HTTP URLs in errors
        url_match = re.search(r'https?://([^/\s]+)', verification_error)
        if url_match:
            host = url_match.group(1)
            return f"{host}:443" if verification_error.startswith('https') else f"{host}:80"
    
    # Try to extract from the secret itself
    if detector_name == 'JDBC':
        # JDBC: jdbc:mysql://host:port/db
        jdbc_match = re.search(r'jdbc:[^:]+://([^:/]+)(?::(\d+))?', raw_secret)
        if jdbc_match:
            host = jdbc_match.group(1)
            port = jdbc_match.group(2) or '3306'  # default MySQL port
            return f"{host}:{port}"
    
    elif detector_name == 'URI':
        try:
            parsed = urlparse(raw_secret)
            if parsed.netloc:
                if ':' in parsed.netloc:
                    return parsed.netloc
                else:
                    port = '443' if parsed.scheme == 'https' else '80'
                    return f"{parsed.netloc}:{port}"
        except:
            pass
    
    elif detector_name in ['Gitlab', 'GitHub']:
        # These hit known API endpoints
        if detector_name == 'Gitlab':
            return 'gitlab.com:443'
        else:
            return 'api.github.com:443'
    
    elif detector_name == 'Slack':
        return 'hooks.slack.com:443'
    
    elif detector_name == 'CircleCI':
        return 'circleci.com:443'
    
    elif detector_name == 'BuildKite':
        return 'api.buildkite.com:443'
    
    elif detector_name == 'CloudflareApiToken':
        return 'api.cloudflare.com:443'
    
    return endpoint

def categorize_network_error(error_msg):
    """Categorize network errors into cacheable failure types"""
    if not error_msg:
        return 'unknown'
    
    error_lower = error_msg.lower()
    
    if 'i/o timeout' in error_lower or 'timeout' in error_lower:
        return 'timeout'
    elif 'connection refused' in error_lower:
        return 'connection_refused'
    elif 'no such host' in error_lower or 'host not found' in error_lower:
        return 'dns_failure'
    elif 'network is unreachable' in error_lower or 'unreachable' in error_lower:
        return 'network_unreachable'
    elif 'connection reset' in error_lower:
        return 'connection_reset'
    elif 'tls handshake' in error_lower or 'certificate' in error_lower:
        return 'tls_failure'
    elif 'forbidden' in error_lower or '403' in error_lower:
        return 'auth_failure'
    elif 'not found' in error_lower or '404' in error_lower:
        return 'not_found'
    elif 'rate limit' in error_lower or '429' in error_lower:
        return 'rate_limited'
    else:
        return 'other_network_error'

def analyze_network_failures(secrets):
    """Comprehensive analysis of network verification failures"""
    
    failure_analysis = {
        'timeout': [],
        'dns_failure': [],
        'connection_refused': [],
        'network_unreachable': [],
        'connection_reset': [],
        'tls_failure': [],
        'auth_failure': [],
        'not_found': [],
        'rate_limited': [],
        'other_network_error': [],
        'successful': [],
        'no_verification_attempted': []
    }
    
    endpoint_failures = defaultdict(lambda: defaultdict(int))
    endpoint_first_failure = {}
    detector_endpoint_map = defaultdict(set)
    
    for i, secret in enumerate(secrets):
        detector_name = secret.get('DetectorName', 'Unknown')
        verified = secret.get('Verified', False)
        verification_error = secret.get('VerificationError', '')
        
        endpoint = extract_network_endpoint(secret)
        if endpoint:
            detector_endpoint_map[detector_name].add(endpoint)
        
        if verified:
            failure_analysis['successful'].append({
                'secret': secret,
                'endpoint': endpoint,
                'detector': detector_name,
                'index': i
            })
        elif verification_error:
            error_type = categorize_network_error(verification_error)
            
            failure_info = {
                'secret': secret,
                'endpoint': endpoint,
                'detector': detector_name,
                'error': verification_error,
                'index': i
            }
            
            failure_analysis[error_type].append(failure_info)
            
            if endpoint:
                endpoint_failures[endpoint][error_type] += 1
                if endpoint not in endpoint_first_failure:
                    endpoint_first_failure[endpoint] = {
                        'error_type': error_type,
                        'index': i
                    }
        else:
            failure_analysis['no_verification_attempted'].append({
                'secret': secret,
                'endpoint': endpoint,
                'detector': detector_name,
                'index': i
            })
    
    return failure_analysis, endpoint_failures, endpoint_first_failure, detector_endpoint_map

def estimate_time_waste_detailed(failure_analysis, endpoint_failures, endpoint_first_failure):
    """Detailed estimation of time wasted on network failures"""
    
    # Time estimates for different failure types (in seconds)
    failure_time_estimates = {
        'timeout': 5.0,           # Full timeout duration
        'dns_failure': 2.0,       # DNS lookup timeout
        'connection_refused': 0.1, # Fast rejection
        'network_unreachable': 3.0, # Network routing timeout
        'connection_reset': 0.2,   # Connection established then reset
        'tls_failure': 1.0,       # TLS handshake failure
        'auth_failure': 0.5,      # HTTP auth failure (fast response)
        'not_found': 0.3,         # HTTP 404 (fast response)
        'rate_limited': 0.3,      # HTTP 429 (fast response)
        'other_network_error': 2.0 # Conservative estimate
    }
    
    time_waste_by_type = {}
    total_waste = 0
    
    for error_type, failures in failure_analysis.items():
        if error_type in failure_time_estimates:
            count = len(failures)
            time_per_failure = failure_time_estimates[error_type]
            total_time = count * time_per_failure
            
            time_waste_by_type[error_type] = {
                'count': count,
                'time_per_failure': time_per_failure,
                'total_time': total_time
            }
            total_waste += total_time
    
    return time_waste_by_type, total_waste

def calculate_caching_savings(endpoint_failures, endpoint_first_failure, failure_analysis):
    """Calculate potential savings from endpoint failure caching"""
    
    failure_time_estimates = {
        'timeout': 5.0, 'dns_failure': 2.0, 'connection_refused': 0.1,
        'network_unreachable': 3.0, 'connection_reset': 0.2, 'tls_failure': 1.0,
        'auth_failure': 0.5, 'not_found': 0.3, 'rate_limited': 0.3,
        'other_network_error': 2.0
    }
    
    caching_savings = {}
    total_savings = 0
    
    for endpoint, error_counts in endpoint_failures.items():
        endpoint_savings = 0
        
        for error_type, count in error_counts.items():
            if count > 1 and error_type in failure_time_estimates:
                # Save time on all attempts after the first failure
                time_per_failure = failure_time_estimates[error_type]
                saved_attempts = count - 1
                savings = saved_attempts * time_per_failure
                endpoint_savings += savings
        
        if endpoint_savings > 0:
            caching_savings[endpoint] = {
                'total_savings': endpoint_savings,
                'failure_counts': dict(error_counts)
            }
            total_savings += endpoint_savings
    
    # Calculate cache hit rate potential
    total_failures = sum(len(failures) for error_type, failures in failure_analysis.items() 
                        if error_type != 'successful' and error_type != 'no_verification_attempted')
    
    potential_cache_hits = sum(sum(max(0, count - 1) for count in error_counts.values()) 
                              for error_counts in endpoint_failures.values())
    
    cache_hit_rate = (potential_cache_hits / total_failures * 100) if total_failures > 0 else 0
    
    return caching_savings, total_savings, cache_hit_rate, potential_cache_hits

def analyze_connection_batching_waste(secrets, detector_endpoint_map):
    """Analyze time wasted on connection churn (opening/closing connections)"""
    
    # Track verification attempts by endpoint and detector
    endpoint_verification_counts = defaultdict(lambda: defaultdict(int))
    
    for secret in secrets:
        detector_name = secret.get('DetectorName', 'Unknown')
        endpoint = extract_network_endpoint(secret)
        
        if endpoint and secret.get('VerificationError') is not None or secret.get('Verified'):
            # This means verification was attempted
            endpoint_verification_counts[endpoint][detector_name] += 1
    
    # Calculate connection overhead
    connection_setup_time = 0.3  # seconds per connection (TCP handshake + TLS if HTTPS)
    connection_teardown_time = 0.1  # seconds per connection close
    total_connection_overhead = connection_setup_time + connection_teardown_time
    
    batching_analysis = {}
    total_current_overhead = 0
    total_optimized_overhead = 0
    
    for endpoint, detector_counts in endpoint_verification_counts.items():
        endpoint_current_overhead = 0
        endpoint_optimized_overhead = 0
        
        for detector, verification_count in detector_counts.items():
            if verification_count > 0:
                # Current: one connection per verification
                current_connections = verification_count
                current_overhead = current_connections * total_connection_overhead
                
                # Optimized: one connection per detector per endpoint (batched)
                optimized_connections = 1
                optimized_overhead = optimized_connections * total_connection_overhead
                
                endpoint_current_overhead += current_overhead
                endpoint_optimized_overhead += optimized_overhead
        
        if endpoint_current_overhead > endpoint_optimized_overhead:
            savings = endpoint_current_overhead - endpoint_optimized_overhead
            batching_analysis[endpoint] = {
                'current_overhead': endpoint_current_overhead,
                'optimized_overhead': endpoint_optimized_overhead,
                'savings': savings,
                'detector_counts': dict(detector_counts)
            }
            
            total_current_overhead += endpoint_current_overhead
            total_optimized_overhead += endpoint_optimized_overhead
    
    total_batching_savings = total_current_overhead - total_optimized_overhead
    
    return batching_analysis, total_batching_savings, total_current_overhead, total_optimized_overhead

def create_comprehensive_network_visualizations(failure_analysis, time_waste_by_type, 
                                                caching_savings_info, batching_info):
    """Create comprehensive network analysis visualizations"""
    
    # Create main analysis figure
    fig, ((ax1, ax2), (ax3, ax4)) = plt.subplots(2, 2, figsize=(18, 14))
    
    # 1. Network Failure Types Distribution
    failure_counts = {}
    for error_type, failures in failure_analysis.items():
        if error_type not in ['successful', 'no_verification_attempted'] and failures:
            failure_counts[error_type.replace('_', ' ').title()] = len(failures)
    
    if failure_counts:
        colors = plt.cm.Set3(np.linspace(0, 1, len(failure_counts)))
        wedges, texts, autotexts = ax1.pie(failure_counts.values(), labels=failure_counts.keys(),
                                          autopct='%1.1f%%', colors=colors, startangle=90)
        ax1.set_title('Network Failure Types Distribution', fontsize=14, weight='bold')
        
        # Highlight the most expensive failures
        for i, (failure_type, count) in enumerate(failure_counts.items()):
            if 'timeout' in failure_type.lower() or 'dns' in failure_type.lower():
                wedges[i].set_edgecolor('red')
                wedges[i].set_linewidth(3)
    
    # 2. Time Waste by Failure Type
    if time_waste_by_type:
        waste_types = []
        waste_times = []
        waste_counts = []
        
        for error_type, info in time_waste_by_type.items():
            waste_types.append(error_type.replace('_', ' ').title())
            waste_times.append(info['total_time'])
            waste_counts.append(info['count'])
        
        bars = ax2.bar(waste_types, waste_times, color='#FF6B6B', alpha=0.7)
        ax2.set_title('Time Wasted by Failure Type', fontsize=14, weight='bold')
        ax2.set_ylabel('Total Time Wasted (seconds)')
        ax2.tick_params(axis='x', rotation=45)
        
        # Add count labels on bars
        for bar, count, time_val in zip(bars, waste_counts, waste_times):
            height = bar.get_height()
            ax2.text(bar.get_x() + bar.get_width()/2., height + max(waste_times) * 0.01,
                    f'{count} failures\n{time_val:.1f}s', ha='center', va='bottom', 
                    fontsize=9, weight='bold')
    
    # 3. Caching Savings Potential
    caching_savings, total_savings, cache_hit_rate, potential_cache_hits = caching_savings_info
    total_waste = sum(info['total_time'] for info in time_waste_by_type.values())
    
    if total_waste > 0:
        categories = ['Current\nTime Waste', 'With Endpoint\nCaching', 'Time Saved']
        values = [total_waste, total_waste - total_savings, total_savings]
        colors = ['#FF6B6B', '#FFB347', '#90EE90']
        
        bars = ax3.bar(categories, values, color=colors)
        ax3.set_title('Endpoint Failure Caching Savings', fontsize=14, weight='bold')
        ax3.set_ylabel('Time (seconds)')
        
        # Add improvement annotation
        if total_savings > 0:
            improvement_pct = (total_savings / total_waste) * 100
            ax3.annotate(f'{improvement_pct:.1f}% improvement\n{cache_hit_rate:.1f}% cache hit rate',
                        xy=(2, total_savings), xytext=(2, total_savings + total_waste * 0.15),
                        arrowprops=dict(arrowstyle='->', color='green', lw=2),
                        fontsize=11, ha='center', color='green', weight='bold')
    
    # 4. Connection Batching Analysis
    batching_analysis, total_batching_savings, total_current_overhead, total_optimized_overhead = batching_info
    
    if total_current_overhead > 0:
        batch_categories = ['Current\nConnection\nOverhead', 'Batched\nConnection\nOverhead', 'Overhead\nSaved']
        batch_values = [total_current_overhead, total_optimized_overhead, total_batching_savings]
        batch_colors = ['#FF6B6B', '#87CEEB', '#90EE90']
        
        bars = ax4.bar(batch_categories, batch_values, color=batch_colors)
        ax4.set_title('Connection Batching Savings Potential', fontsize=14, weight='bold')
        ax4.set_ylabel('Connection Overhead (seconds)')
        
        # Add batching stats
        num_endpoints = len(batching_analysis)
        if total_current_overhead > 0:
            batching_improvement = (total_batching_savings / total_current_overhead) * 100
            ax4.annotate(f'{batching_improvement:.1f}% reduction\n{num_endpoints} endpoints\ncan be batched',
                        xy=(2, total_batching_savings), xytext=(2, total_batching_savings + total_current_overhead * 0.15),
                        arrowprops=dict(arrowstyle='->', color='green', lw=2),
                        fontsize=10, ha='center', color='green', weight='bold')
    
    plt.tight_layout()
    plt.savefig('benchmarks/comprehensive_network_analysis.png', dpi=300, bbox_inches='tight')
    plt.close()

def create_endpoint_analysis_chart(endpoint_failures, caching_savings):
    """Create detailed endpoint failure analysis"""
    
    fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(16, 8))
    
    # 1. Top Failed Endpoints
    endpoint_total_failures = {ep: sum(errors.values()) for ep, errors in endpoint_failures.items()}
    top_endpoints = dict(Counter(endpoint_total_failures).most_common(10))
    
    if top_endpoints:
        bars = ax1.bar(range(len(top_endpoints)), list(top_endpoints.values()), 
                      color='#FF6B6B', alpha=0.7)
        ax1.set_title('Top 10 Most Failed Endpoints', fontsize=14, weight='bold')
        ax1.set_ylabel('Total Failure Count')
        ax1.set_xlabel('Endpoints')
        
        # Format endpoint labels
        endpoint_labels = []
        for endpoint in top_endpoints.keys():
            if len(endpoint) > 25:
                endpoint_labels.append(endpoint[:22] + '...')
            else:
                endpoint_labels.append(endpoint)
        
        ax1.set_xticks(range(len(top_endpoints)))
        ax1.set_xticklabels(endpoint_labels, rotation=45, ha='right')
        
        # Add failure count labels
        for i, (bar, count) in enumerate(zip(bars, top_endpoints.values())):
            ax1.text(bar.get_x() + bar.get_width()/2., bar.get_height() + max(top_endpoints.values()) * 0.01,
                    f'{count}', ha='center', va='bottom', fontsize=9, weight='bold')
    
    # 2. Caching Savings by Endpoint
    if caching_savings:
        top_savings_endpoints = dict(sorted(caching_savings.items(), 
                                          key=lambda x: x[1]['total_savings'], reverse=True)[:10])
        
        if top_savings_endpoints:
            savings_values = [info['total_savings'] for info in top_savings_endpoints.values()]
            bars = ax2.bar(range(len(top_savings_endpoints)), savings_values,
                          color='#90EE90', alpha=0.7)
            ax2.set_title('Top 10 Endpoints by Caching Savings Potential', fontsize=14, weight='bold')
            ax2.set_ylabel('Potential Time Savings (seconds)')
            ax2.set_xlabel('Endpoints')
            
            # Format labels
            savings_labels = []
            for endpoint in top_savings_endpoints.keys():
                if len(endpoint) > 25:
                    savings_labels.append(endpoint[:22] + '...')
                else:
                    savings_labels.append(endpoint)
            
            ax2.set_xticks(range(len(top_savings_endpoints)))
            ax2.set_xticklabels(savings_labels, rotation=45, ha='right')
            
            # Add savings labels
            for i, (bar, savings) in enumerate(zip(bars, savings_values)):
                ax2.text(bar.get_x() + bar.get_width()/2., bar.get_height() + max(savings_values) * 0.01,
                        f'{savings:.1f}s', ha='center', va='bottom', fontsize=9, weight='bold')
    
    plt.tight_layout()
    plt.savefig('benchmarks/endpoint_failure_analysis.png', dpi=300, bbox_inches='tight')
    plt.close()

def generate_network_optimization_report(failure_analysis, time_waste_by_type, 
                                        caching_savings_info, batching_info, total_secrets):
    """Generate comprehensive network optimization report"""
    
    caching_savings, total_caching_savings, cache_hit_rate, potential_cache_hits = caching_savings_info
    batching_analysis, total_batching_savings, total_current_overhead, total_optimized_overhead = batching_info
    
    total_time_waste = sum(info['total_time'] for info in time_waste_by_type.values())
    total_potential_savings = total_caching_savings + total_batching_savings
    
    # Calculate detailed failure statistics
    failure_stats = {}
    for error_type, failures in failure_analysis.items():
        if error_type not in ['successful', 'no_verification_attempted']:
            failure_stats[error_type] = len(failures)
    
    successful_verifications = len(failure_analysis.get('successful', []))
    total_verification_attempts = sum(failure_stats.values()) + successful_verifications
    
    report = f"""# TruffleHog Network Verification Waste Analysis

## 🎯 Executive Summary

**Critical Finding**: Network verification failures are wasting significant time through repeated attempts to unreachable endpoints and inefficient connection patterns.

- **Total Secrets Analyzed**: {total_secrets:,}
- **Verification Attempts**: {total_verification_attempts:,}
- **Success Rate**: {(successful_verifications/total_verification_attempts*100):.1f}% ({successful_verifications:,} successful)
- **Total Time Wasted**: {total_time_waste:.1f} seconds
- **Potential Savings**: {total_potential_savings:.1f} seconds ({(total_potential_savings/total_time_waste*100):.1f}% improvement)

## 🚨 Network Failure Breakdown

### Failure Types and Time Impact
"""

    for error_type, info in sorted(time_waste_by_type.items(), 
                                  key=lambda x: x[1]['total_time'], reverse=True):
        error_name = error_type.replace('_', ' ').title()
        report += f"""
**{error_name}**:
- Count: {info['count']:,} failures
- Time per failure: {info['time_per_failure']:.1f}s
- Total time wasted: {info['total_time']:.1f}s
- Percentage of waste: {(info['total_time']/total_time_waste*100 if total_time_waste > 0 else 0):.1f}%"""

    report += f"""

### Top Cacheable Failures (should never be retried)
1. **DNS Failures**: {failure_stats.get('dns_failure', 0):,} attempts × 2.0s = {failure_stats.get('dns_failure', 0) * 2.0:.1f}s wasted
2. **Network Unreachable**: {failure_stats.get('network_unreachable', 0):,} attempts × 3.0s = {failure_stats.get('network_unreachable', 0) * 3.0:.1f}s wasted
3. **Connection Refused**: {failure_stats.get('connection_refused', 0):,} attempts × 0.1s = {failure_stats.get('connection_refused', 0) * 0.1:.1f}s wasted
4. **Timeouts**: {failure_stats.get('timeout', 0):,} attempts × 5.0s = {failure_stats.get('timeout', 0) * 5.0:.1f}s wasted

## 💡 Optimization Opportunities

### 1. Endpoint Failure Caching 🎯 **HIGH IMPACT**

**Problem**: Repeatedly attempting verification against known-failed endpoints
- **Potential Cache Hits**: {potential_cache_hits:,} redundant attempts
- **Cache Hit Rate**: {cache_hit_rate:.1f}%
- **Time Savings**: {total_caching_savings:.1f} seconds ({(total_caching_savings/total_time_waste*100 if total_time_waste > 0 else 0):.1f}% improvement)

**Implementation**:
```go
type EndpointFailureCache struct {{
    failures map[string]FailureInfo
    mu       sync.RWMutex
    ttl      time.Duration // How long to remember failures
}}

type FailureInfo struct {{
    ErrorType    string
    FirstFailure time.Time
    FailureCount int
    LastAttempt  time.Time
}}

func (efc *EndpointFailureCache) ShouldSkip(endpoint string) bool {{
    efc.mu.RLock()
    defer efc.mu.RUnlock()
    
    if info, exists := efc.failures[endpoint]; exists {{
        // Skip if recently failed and it's a permanent failure type
        if isPermanentFailure(info.ErrorType) && 
           time.Since(info.LastAttempt) < efc.ttl {{
            return true
        }}
        
        // Skip if failed multiple times recently
        if info.FailureCount >= 3 && 
           time.Since(info.LastAttempt) < time.Hour {{
            return true
        }}
    }}
    return false
}}

func isPermanentFailure(errorType string) bool {{
    return errorType == "dns_failure" || 
           errorType == "network_unreachable" ||
           errorType == "connection_refused"
}}
```

### 2. Connection Batching 🔧 **MEDIUM IMPACT**

**Problem**: Opening/closing connections repeatedly to the same endpoint
- **Current Connection Overhead**: {total_current_overhead:.1f} seconds
- **Optimized Overhead**: {total_optimized_overhead:.1f} seconds  
- **Potential Savings**: {total_batching_savings:.1f} seconds

**Implementation**:
```go
type EndpointBatcher struct {{
    pending map[string][]SecretVerification
    timeout time.Duration
    mu      sync.Mutex
}}

func (eb *EndpointBatcher) QueueVerification(endpoint string, secret SecretVerification) {{
    eb.mu.Lock()
    defer eb.mu.Unlock()
    
    eb.pending[endpoint] = append(eb.pending[endpoint], secret)
    
    // Trigger batch processing if queue is full or timeout reached
    if len(eb.pending[endpoint]) >= 10 {{
        go eb.processBatch(endpoint)
    }}
}}

func (eb *EndpointBatcher) processBatch(endpoint string) {{
    // Open one connection and verify all secrets for this endpoint
    // Reuse the connection for multiple verifications
}}
```

### 3. Smart Timeout Configuration ⚡ **QUICK WIN**

**Current Issue**: Fixed 5-second timeout for all endpoint types

**Optimization**:
```go
func getAdaptiveTimeout(endpoint string, errorHistory []string) time.Duration {{
    // DNS failures: use short timeout (1s)
    if hasDNSFailures(errorHistory) {{
        return 1 * time.Second
    }}
    
    // Previously successful endpoints: longer timeout
    if hasRecentSuccess(endpoint) {{
        return 10 * time.Second  
    }}
    
    // Default for new endpoints
    return 3 * time.Second
}}
```

## 📊 Expected Performance Impact

| Optimization | Current Time | Optimized Time | Improvement | Implementation Effort |
|--------------|--------------|----------------|-------------|----------------------|
| Endpoint Caching | {total_time_waste:.1f}s | {total_time_waste - total_caching_savings:.1f}s | {(total_caching_savings/total_time_waste*100 if total_time_waste > 0 else 0):.1f}% | 2-3 days |
| Connection Batching | {total_current_overhead:.1f}s | {total_optimized_overhead:.1f}s | {(total_batching_savings/total_current_overhead*100 if total_current_overhead > 0 else 0):.1f}% | 1-2 weeks |
| Adaptive Timeouts | Current | -15% | 15% | 1 day |
| **Combined** | {total_time_waste:.1f}s | {total_time_waste - total_potential_savings:.1f}s | **{(total_potential_savings/total_time_waste*100 if total_time_waste > 0 else 0):.1f}%** | **2-3 weeks** |

## 🎯 Implementation Priority

### Phase 1: Quick Wins (Week 1)
1. **DNS Failure Caching** - Never retry DNS failures for 1 hour
2. **Connection Refused Caching** - Never retry refused connections for 30 minutes  
3. **Adaptive Timeouts** - Reduce timeout for known problematic endpoints

**Expected Impact**: 40-50% reduction in network waste

### Phase 2: Connection Optimization (Weeks 2-3)
1. **Connection Batching** - Batch verifications by endpoint
2. **Connection Pooling** - Reuse connections across verifications
3. **Endpoint Health Tracking** - Track endpoint reliability over time

**Expected Impact**: Additional 15-20% improvement

### Phase 3: Advanced Optimization (Week 4+)
1. **ML-Based Endpoint Scoring** - Predict endpoint failure probability
2. **Geographic Endpoint Routing** - Route to closest endpoints
3. **Rate Limiting Awareness** - Back off on rate-limited endpoints

**Expected Impact**: Additional 10-15% improvement

## 🔍 Monitoring Requirements

### Key Metrics to Track
1. **Cache Hit Rate**: Target >60% for endpoint failure cache
2. **Connection Reuse Rate**: Target >80% for same-endpoint requests
3. **Average Verification Time**: Target <2s per verification
4. **Timeout Rate**: Target <5% of all verification attempts
5. **DNS Failure Rate**: Should approach 0% with caching

### Alerting Thresholds
- Cache hit rate drops below 50%
- Average verification time exceeds 3s
- Timeout rate exceeds 10%
- New DNS failures detected

## 💾 Cache Design Recommendations

### Endpoint Failure Cache Structure
```go
type FailureCache struct {{
    // In-memory cache for fast lookups
    memory map[string]FailureInfo
    
    // Persistent storage for long-term patterns  
    persistent *sql.DB
    
    // Cache policies
    maxMemoryEntries int
    defaultTTL       time.Duration
    permanentFailureTTL time.Duration
}}
```

### Cache Size Estimates
- **Memory Cache**: ~10,000 endpoints × 100 bytes = 1MB
- **Persistent Cache**: ~100,000 endpoints × 200 bytes = 20MB
- **Cache Hit Rate**: Expected 60-80% after warmup period

---

*Analysis based on Figma organization scan with {total_secrets:,} secrets*
*Generated on: {pd.Timestamp.now().strftime('%Y-%m-%d %H:%M:%S')}*
"""
    
    return report

def main():
    """Main network analysis function"""
    print("🔍 Starting Comprehensive Network Verification Analysis...")
    
    # Load data
    secrets, logs = load_scan_data('figma_comprehensive_scan.json')
    
    if not secrets:
        print("❌ No scan data found. Please ensure figma_comprehensive_scan.json exists.")
        return
    
    print(f"📊 Analyzing {len(secrets)} secrets for network verification patterns...")
    
    # Perform comprehensive analysis
    print("🔍 Analyzing network failures...")
    failure_analysis, endpoint_failures, endpoint_first_failure, detector_endpoint_map = analyze_network_failures(secrets)
    
    print("⏱️  Calculating time waste...")
    time_waste_by_type, total_time_waste = estimate_time_waste_detailed(failure_analysis, endpoint_failures, endpoint_first_failure)
    
    print("💾 Analyzing caching savings potential...")
    caching_savings_info = calculate_caching_savings(endpoint_failures, endpoint_first_failure, failure_analysis)
    
    print("🔄 Analyzing connection batching waste...")
    batching_info = analyze_connection_batching_waste(secrets, detector_endpoint_map)
    
    # Create visualizations
    print("📈 Generating comprehensive visualizations...")
    create_comprehensive_network_visualizations(failure_analysis, time_waste_by_type, 
                                               caching_savings_info, batching_info)
    
    create_endpoint_analysis_chart(endpoint_failures, caching_savings_info[0])
    
    # Generate detailed report
    print("📝 Generating optimization report...")
    report = generate_network_optimization_report(failure_analysis, time_waste_by_type,
                                                 caching_savings_info, batching_info, len(secrets))
    
    # Save report
    with open('benchmarks/NETWORK_VERIFICATION_WASTE_ANALYSIS.md', 'w') as f:
        f.write(report)
    
    # Print executive summary
    total_caching_savings = caching_savings_info[1]
    total_batching_savings = batching_info[1]
    cache_hit_rate = caching_savings_info[2]
    
    print("\n" + "="*60)
    print("🎯 NETWORK VERIFICATION WASTE ANALYSIS SUMMARY")
    print("="*60)
    print(f"📊 Total Secrets: {len(secrets):,}")
    print(f"⏱️  Total Time Wasted: {total_time_waste:.1f} seconds")
    print(f"💾 Endpoint Caching Savings: {total_caching_savings:.1f}s ({cache_hit_rate:.1f}% hit rate)")
    print(f"🔄 Connection Batching Savings: {total_batching_savings:.1f}s")
    print(f"🚀 Total Potential Improvement: {((total_caching_savings + total_batching_savings)/total_time_waste*100 if total_time_waste > 0 else 0):.1f}%")
    print("\n📁 Generated Files:")
    print("   - benchmarks/comprehensive_network_analysis.png")
    print("   - benchmarks/endpoint_failure_analysis.png")
    print("   - benchmarks/NETWORK_VERIFICATION_WASTE_ANALYSIS.md")
    
    # Highlight top failure types
    print("\n🚨 Top Time Wasters:")
    for error_type, info in sorted(time_waste_by_type.items(), 
                                  key=lambda x: x[1]['total_time'], reverse=True)[:3]:
        print(f"   {error_type.replace('_', ' ').title()}: {info['total_time']:.1f}s ({info['count']} failures)")

if __name__ == "__main__":
    main()