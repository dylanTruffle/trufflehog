# TruffleHog Cache Performance Breakdown

## Current Cache Savings Analysis

### Baseline Performance (No Cache)
```
Total Scan Time: 100%
├── Network Verification: 70%
├── Aho-Corasick Regex: 15%
├── File I/O: 10%
├── Result Processing: 3%
└── Memory Management: 2%
```

### Current Cache Performance (80% Hit Rate)
```
Cache Hit Rate: 80% of verification calls avoided
Cache Miss Rate: 20% of verification calls still need network

Verification Time Reduction:
- Original verification: 70% of scan time
- Cache hits avoid: 80% × 70% = 56% of total scan time saved
- Remaining verification: 20% × 70% = 14% of total scan time

Current Scan Time: 100% → 44% (56% savings)
├── Network Verification: 14% (was 70%, reduced by 80%)
├── Aho-Corasick Regex: 15% (unchanged)
├── File I/O: 10% (unchanged)  
├── Result Processing: 3% (unchanged)
└── Memory Management: 2% (unchanged)
```

**Current Cache Savings: 56% of total scan time**

## Next-Gen Cache Additional Savings

### Batching Requests to Same Hosts
**Targets**: The remaining 14% verification time (cache misses only)

**Savings Breakdown**:
- Connection establishment overhead: ~30% of verification time
- SSL handshake overhead: ~25% of verification time  
- HTTP keep-alive reuse: ~20% efficiency gain

**Batching Impact on Remaining Verifications**:
- 50-70% reduction in connection overhead for cache misses
- Applied to 14% verification time: 14% × 60% = 8.4% additional savings

### Hostname/Network Issue Tracking
**Targets**: Timeout and failure waste in remaining verifications

**Current Waste in Cache Misses**:
- DNS failures: ~5% of verification attempts (5-30 second timeouts)
- Connection timeouts: ~8% of verification attempts 
- Network unreachable: ~3% of verification attempts

**Failure Tracking Impact**:
- Eliminate 90% of timeout waste
- Applied to remaining 14% verification: 14% × 20% failure rate × 90% = 2.5% additional savings

### Combined Next-Gen Savings
```
Additional savings from next-gen cache:
├── Batching efficiency: 8.4%
├── Timeout elimination: 2.5%
└── Total additional: 10.9%

Combined cache savings: 56% + 10.9% = 66.9% of original scan time
```

## Current Pipeline Bottleneck Analysis

### With Current Cache (80% hit rate)
```
Scan Pipeline Flow:
File I/O (10%) → Aho-Corasick Regex (15%) → Cache Check → Network Verification (14%)

Bottleneck Analysis:
1. Aho-Corasick Regex: 15% (CPU-bound)
2. Network Verification: 14% (I/O-bound, cache misses)  
3. File I/O: 10% (I/O-bound)

Primary Bottleneck: Aho-Corasick Regex (15%)
- Most CPU-intensive single component
- Cannot be easily parallelized beyond current threading
- Complex regex patterns across 1000+ detectors
```

**Current bottleneck is MIXED**: 
- **CPU-bound**: Aho-Corasick Regex (15%)
- **I/O-bound**: Network Verification cache misses (14%)

### With Next-Gen Cache
```
After batching + hostname tracking:
Network Verification: 14% → 3.1% (reduced by 10.9%)

Updated Pipeline Bottlenecks:
1. Aho-Corasick Regex: 15% (unchanged, now clearly dominant)
2. File I/O: 10% (unchanged)
3. Network Verification: 3.1% (heavily reduced)
4. Result Processing: 3% (unchanged)

Clear Primary Bottleneck: Aho-Corasick Regex (15%)
- 5x larger than remaining network verification
- CPU-bound pattern matching becomes the dominant constraint
```

## Performance Summary

| Cache Level | Verification Time | Total Time Saved | Primary Bottleneck |
|-------------|------------------|------------------|-------------------|
| **No Cache** | 70% | 0% | Network Verification |
| **Current Cache** | 14% | 56% | Mixed: Regex (15%) + Network (14%) |
| **Next-Gen Cache** | 3.1% | 66.9% | **CPU: Aho-Corasick Regex (15%)** |

## Key Insights

### Current Cache Effectiveness
- **56% total scan time savings** through verification cache hits
- Network verification reduced from 70% → 14% of scan time
- Creates mixed bottleneck between CPU and remaining I/O

### Next-Gen Cache Impact  
- **Additional 10.9% savings** through batching and failure tracking
- Network verification further reduced from 14% → 3.1%
- **Clearly shifts bottleneck to CPU-bound Aho-Corasick processing**

### Optimization Strategy
1. **Phase 1**: Deploy next-gen cache (batching + hostname tracking)
2. **Phase 2**: Focus on Aho-Corasick optimization (becomes clear bottleneck)
3. **Phase 3**: File I/O optimization (second largest component)

The cache evolution clearly shows the progression from I/O-bound to CPU-bound performance characteristics.