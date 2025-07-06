#!/usr/bin/env python3
"""
Comprehensive Analysis of TruffleHog Figma Scan
Analyzes performance bottlenecks and pipeline metrics
"""

import json
import sys
from collections import defaultdict, Counter
from datetime import datetime
import statistics
import re

def analyze_scan_results(filename):
    """Parse and analyze TruffleHog scan results"""
    
    print("🔍 TruffleHog Figma Scan Analysis")
    print("=" * 50)
    
    try:
        with open(filename, 'r') as f:
            lines = f.readlines()
    except FileNotFoundError:
        print(f"❌ File {filename} not found")
        return
    
    # Parse results
    secrets = []
    logs = []
    
    for line in lines:
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
    
    print(f"📊 Total secrets found: {len(secrets)}")
    print(f"📋 Total log entries: {len(logs)}")
    print()
    
    # Analyze secrets by type
    analyze_secrets(secrets)
    
    # Analyze repositories
    analyze_repositories(secrets)
    
    # Analyze performance patterns
    analyze_performance(secrets, logs)
    
    # Repository size analysis
    analyze_repository_sizes(secrets)
    
    # Verification analysis
    analyze_verification(secrets)
    
    # Generate recommendations
    generate_recommendations(secrets, logs)

def analyze_secrets(secrets):
    """Analyze secrets by detector type"""
    print("🔐 Secret Analysis by Detector Type")
    print("-" * 40)
    
    detector_counts = Counter()
    verification_stats = defaultdict(list)
    
    for secret in secrets:
        detector = secret.get('DetectorName', 'Unknown')
        detector_counts[detector] += 1
        
        verified = secret.get('Verified', False)
        verification_stats[detector].append(verified)
    
    # Top detectors
    print("Top 10 Detector Types:")
    for detector, count in detector_counts.most_common(10):
        verified_count = sum(verification_stats[detector])
        verification_rate = (verified_count / count) * 100 if count > 0 else 0
        print(f"  {detector:20} {count:3} secrets ({verification_rate:.1f}% verified)")
    
    print()

def analyze_repositories(secrets):
    """Analyze secrets by repository"""
    print("📁 Repository Analysis")
    print("-" * 40)
    
    repo_stats = defaultdict(lambda: {
        'total_secrets': 0,
        'verified_secrets': 0,
        'detectors': set(),
        'files': set()
    })
    
    for secret in secrets:
        repo_url = secret.get('SourceMetadata', {}).get('Data', {}).get('Github', {}).get('repository', 'Unknown')
        repo_name = repo_url.split('/')[-1].replace('.git', '') if repo_url != 'Unknown' else 'Unknown'
        
        repo_stats[repo_name]['total_secrets'] += 1
        if secret.get('Verified', False):
            repo_stats[repo_name]['verified_secrets'] += 1
        
        detector = secret.get('DetectorName', 'Unknown')
        repo_stats[repo_name]['detectors'].add(detector)
        
        file_path = secret.get('SourceMetadata', {}).get('Data', {}).get('Github', {}).get('file', 'Unknown')
        repo_stats[repo_name]['files'].add(file_path)
    
    # Sort by total secrets
    sorted_repos = sorted(repo_stats.items(), key=lambda x: x[1]['total_secrets'], reverse=True)
    
    print("Top 10 Repositories by Secret Count:")
    for repo, stats in sorted_repos[:10]:
        print(f"  {repo:30} {stats['total_secrets']:3} secrets, {stats['verified_secrets']:2} verified")
        print(f"    {'':32} {len(stats['detectors'])} detector types, {len(stats['files'])} files")
    
    print()

def analyze_performance(secrets, logs):
    """Analyze performance patterns"""
    print("⚡ Performance Analysis")
    print("-" * 40)
    
    # Analyze scan timeline
    if logs:
        print("📈 Scan Timeline Analysis:")
        scan_messages = [log for log in logs if 'scanning repo' in log.get('msg', '')]
        print(f"  Repositories scanned: {len(scan_messages)}")
        
        # Find unique repositories from logs
        repos_from_logs = set()
        for log in scan_messages:
            repo_url = log.get('repo', '')
            if repo_url:
                repo_name = repo_url.split('/')[-1].replace('.git', '')
                repos_from_logs.add(repo_name)
        
        print(f"  Unique repositories from logs: {len(repos_from_logs)}")
    
    # Analyze decoders
    decoder_stats = Counter()
    for secret in secrets:
        decoder = secret.get('DecoderName', 'Unknown')
        decoder_stats[decoder] += 1
    
    print("\n🔧 Decoder Usage:")
    for decoder, count in decoder_stats.most_common():
        percentage = (count / len(secrets)) * 100 if secrets else 0
        print(f"  {decoder:15} {count:3} secrets ({percentage:.1f}%)")
    
    print()

def analyze_repository_sizes(secrets):
    """Analyze repository characteristics"""
    print("📏 Repository Characteristics")
    print("-" * 40)
    
    repo_files = defaultdict(set)
    repo_commits = defaultdict(set)
    
    for secret in secrets:
        github_data = secret.get('SourceMetadata', {}).get('Data', {}).get('Github', {})
        repo_url = github_data.get('repository', '')
        repo_name = repo_url.split('/')[-1].replace('.git', '') if repo_url else 'Unknown'
        
        file_path = github_data.get('file', '')
        commit = github_data.get('commit', '')
        
        if file_path:
            repo_files[repo_name].add(file_path)
        if commit:
            repo_commits[repo_name].add(commit)
    
    print("Repository Activity (Top 10):")
    repo_activity = {}
    for repo in repo_files:
        repo_activity[repo] = {
            'files': len(repo_files[repo]),
            'commits': len(repo_commits[repo])
        }
    
    sorted_activity = sorted(repo_activity.items(), key=lambda x: x[1]['files'], reverse=True)
    
    for repo, activity in sorted_activity[:10]:
        print(f"  {repo:30} {activity['files']:3} files, {activity['commits']:3} commits")
    
    print()

def analyze_verification(secrets):
    """Analyze verification patterns"""
    print("✅ Verification Analysis")
    print("-" * 40)
    
    verified_secrets = [s for s in secrets if s.get('Verified', False)]
    unverified_secrets = [s for s in secrets if not s.get('Verified', False)]
    
    print(f"Verified secrets: {len(verified_secrets)}")
    print(f"Unverified secrets: {len(unverified_secrets)}")
    
    if secrets:
        verification_rate = (len(verified_secrets) / len(secrets)) * 100
        print(f"Overall verification rate: {verification_rate:.1f}%")
    
    # Analyze verification errors
    verification_errors = []
    for secret in secrets:
        if 'VerificationError' in secret:
            verification_errors.append(secret['VerificationError'])
    
    if verification_errors:
        print(f"\nVerification errors: {len(verification_errors)}")
        error_types = Counter(verification_errors)
        for error, count in error_types.most_common(5):
            print(f"  {error[:60]}{'...' if len(error) > 60 else ''}: {count}")
    
    print()

def generate_recommendations(secrets, logs):
    """Generate performance recommendations"""
    print("💡 Performance Recommendations")
    print("-" * 40)
    
    recommendations = []
    
    # Verification analysis
    if secrets:
        verification_rate = len([s for s in secrets if s.get('Verified', False)]) / len(secrets) * 100
        if verification_rate < 50:
            recommendations.append({
                'priority': 'HIGH',
                'category': 'Verification',
                'issue': f'Low verification rate ({verification_rate:.1f}%)',
                'recommendation': 'Network verification is likely the bottleneck. Consider parallel verification or caching.'
            })
    
    # Detector analysis
    detector_counts = Counter()
    for secret in secrets:
        detector_counts[secret.get('DetectorName', 'Unknown')] += 1
    
    if detector_counts:
        top_detector = detector_counts.most_common(1)[0]
        if top_detector[1] > len(secrets) * 0.3:
            recommendations.append({
                'priority': 'MEDIUM',
                'category': 'Detection',
                'issue': f'{top_detector[0]} detector found {top_detector[1]} secrets ({top_detector[1]/len(secrets)*100:.1f}% of total)',
                'recommendation': 'Consider optimizing the most active detector patterns for better performance.'
            })
    
    # Repository analysis
    repo_stats = defaultdict(int)
    for secret in secrets:
        repo_url = secret.get('SourceMetadata', {}).get('Data', {}).get('Github', {}).get('repository', 'Unknown')
        repo_name = repo_url.split('/')[-1].replace('.git', '') if repo_url != 'Unknown' else 'Unknown'
        repo_stats[repo_name] += 1
    
    if repo_stats:
        top_repo = max(repo_stats.items(), key=lambda x: x[1])
        if top_repo[1] > len(secrets) * 0.4:
            recommendations.append({
                'priority': 'LOW',
                'category': 'Repository',
                'issue': f'{top_repo[0]} contains {top_repo[1]} secrets ({top_repo[1]/len(secrets)*100:.1f}% of total)',
                'recommendation': 'Large repositories may benefit from incremental scanning or filtering.'
            })
    
    # Output recommendations
    if recommendations:
        for rec in sorted(recommendations, key=lambda x: {'HIGH': 0, 'MEDIUM': 1, 'LOW': 2}[x['priority']]):
            print(f"[{rec['priority']}] {rec['category']}: {rec['issue']}")
            print(f"    💡 {rec['recommendation']}")
            print()
    else:
        print("✅ No significant performance issues detected.")
        print()
    
    # Summary
    print("📝 Summary")
    print("-" * 40)
    print(f"• Total secrets analyzed: {len(secrets)}")
    print(f"• Repository coverage: {len(set(s.get('SourceMetadata', {}).get('Data', {}).get('Github', {}).get('repository', 'Unknown') for s in secrets))}")
    print(f"• Detector types used: {len(set(s.get('DetectorName', 'Unknown') for s in secrets))}")
    print(f"• Verification rate: {len([s for s in secrets if s.get('Verified', False)]) / len(secrets) * 100:.1f}%" if secrets else "• No secrets to analyze")

if __name__ == "__main__":
    filename = "figma_comprehensive_scan.json"
    if len(sys.argv) > 1:
        filename = sys.argv[1]
    
    analyze_scan_results(filename)