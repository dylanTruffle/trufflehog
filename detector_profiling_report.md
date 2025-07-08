# Detector Profiling Report

This report shows detailed performance metrics for each detector during the scan.

## Summary

- **Total Detectors**: 3
- **Total Execution Time**: 622.88ms
- **Total Detector Calls**: 6
- **Average Time per Call**: 103.813ms

## Top 10 Slowest Detectors (by total time)

1. **PrivateKey**
   - Total Time: 479.048ms
   - Calls: 2
   - Average: 239.524ms
   - Min: 239.103ms
   - Max: 239.946ms

2. **URI**
   - Total Time: 83.186ms
   - Calls: 2
   - Average: 41.593ms
   - Min: 40.58ms
   - Max: 42.606ms

3. **AWS**
   - Total Time: 60.645ms
   - Calls: 2
   - Average: 30.323ms
   - Min: 26.297ms
   - Max: 34.349ms

## Complete Detector Performance Table

| Detector | Calls | Total Time | Average | Min | Max |
|----------|-------|------------|---------|-----|-----|
| PrivateKey | 2 | 479.048ms | 239.524ms | 239.103ms | 239.946ms |
| URI | 2 | 83.186ms | 41.593ms | 40.58ms | 42.606ms |
| AWS | 2 | 60.645ms | 30.323ms | 26.297ms | 34.349ms |

## Analysis

The slowest detector is **PrivateKey** with a total execution time of 479.048ms across 2 calls.
The top 3 detectors account for 100.0% of total detection time.
