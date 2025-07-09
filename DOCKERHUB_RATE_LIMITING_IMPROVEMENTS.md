# DockerHub Rate Limiting Improvements

## Overview
This document describes the comprehensive rate limiting improvements implemented for the DockerHub source in TruffleHog's projectalpha branch.

## Problem Statement
The original DockerHub source implementation detected HTTP 429 rate limit responses but only logged them without implementing any retry or backoff mechanism. This led to scan failures when DockerHub's API rate limits were exceeded, particularly problematic for:

- Anonymous users (100 requests/6 hours)
- Authenticated free tier users (200 requests/6 hours)
- Large organization scans with many repositories and tags

## Solution Implemented

### 1. Comprehensive Rate Limit Detection
- **HTTP 429 Status Code Detection**: Automatically detects when DockerHub returns rate limit responses
- **Retry-After Header Parsing**: Respects RFC 7231 compliant `Retry-After` headers (both seconds and HTTP-date formats)
- **Global State Management**: Coordinates rate limit state across all concurrent goroutines to prevent redundant API calls

### 2. Exponential Backoff with Jitter
- **Base Delay**: 2-minute initial delay for rate limit recovery
- **Maximum Delay**: 30-minute cap to prevent excessive waiting
- **Jitter**: 10-30 second randomization to prevent thundering herd effects
- **Fallback Strategy**: Uses intelligent defaults when `Retry-After` header is missing

### 3. Retry Logic with Context Awareness
- **Configurable Retries**: Maximum 3 retry attempts per request
- **Context Cancellation**: Respects context cancellation during rate limit waits
- **Network Error Handling**: Separate exponential backoff for network errors vs. rate limits
- **Request Retry Wrapper**: `makeRequestWithRetry()` function handles all retry logic

### 4. Enhanced Logging and Observability
- **Structured Logging**: Detailed logs with retry attempts, wait durations, and resume times
- **Rate Limit Tracking**: Logs estimated reset times (6-hour DockerHub cycles)
- **Progress Visibility**: Clear indication when waiting for rate limits vs. actual scanning

## Key Functions Added

### `handleRateLimit(ctx, resp)`
- Detects and handles rate limit responses
- Manages global rate limit state with mutex protection
- Implements intelligent waiting with context cancellation support
- Returns true if rate limit was handled and request should be retried

### `parseRetryAfter(resp)`
- Parses `Retry-After` header in both seconds and HTTP-date formats
- Validates reasonable retry durations (max 1 hour)
- Provides safe fallback when header parsing fails

### `makeRequestWithRetry(ctx, req, maxRetries)`
- Wrapper function that combines rate limiting and network error retry logic
- Handles both temporary network failures and API rate limits
- Ensures proper response body cleanup on retries
- Provides detailed logging for debugging

## Integration Points

### 1. Repository Enumeration
- Updated `getRepositoriesForOrganization()` to use new retry logic
- Replaced direct `httpClient.Do()` calls with `makeRequestWithRetry()`
- Removed old rate limit detection code that just logged and returned

### 2. Tag Discovery
- Updated `getImagesForRepository()` to use new retry logic
- Maintains pagination state across retries
- Ensures tag enumeration continues after rate limit recovery

### 3. Global Coordination
- All DockerHub API requests now coordinate through shared rate limit state
- Prevents multiple goroutines from simultaneously hitting rate limits
- Ensures efficient resource usage during rate limit periods

## Benefits

### 1. Improved Reliability
- **Automatic Recovery**: Scans continue automatically after rate limit periods
- **Reduced Failures**: Eliminates scan failures due to temporary rate limiting
- **Better Resource Usage**: Intelligent backoff prevents wasted API calls

### 2. Enhanced User Experience
- **Clear Progress Indication**: Users know when scans are waiting vs. actively scanning
- **Predictable Behavior**: Consistent handling across different rate limit scenarios
- **Configurable Behavior**: Retry limits can be adjusted based on requirements

### 3. Production Readiness
- **Context Awareness**: Proper cancellation support for long-running scans
- **Memory Efficiency**: No resource leaks during extended rate limit periods
- **Monitoring Support**: Detailed logging for operational visibility

## Configuration Options
The rate limiting behavior can be controlled through:
- **Retry Limits**: Currently set to 3 attempts per request
- **Base Delays**: 2-minute base delay with 30-minute maximum
- **Authentication**: Higher rate limits with DockerHub authentication tokens

## Backward Compatibility
- All existing configuration options remain unchanged
- No breaking changes to the DockerHub source interface
- Maintains compatibility with existing TruffleHog deployment patterns

## Testing Considerations
To test the rate limiting implementation:

1. **Unauthenticated Scans**: Test with anonymous access to trigger rate limits faster
2. **Large Organizations**: Scan organizations with 100+ repositories to exercise retry logic
3. **Network Interruption**: Test context cancellation during rate limit waits
4. **Authentication**: Verify higher rate limits with DockerHub tokens

## Future Enhancements
Potential improvements for future releases:

1. **Metrics Integration**: Add Prometheus metrics for rate limit encounters (similar to GitHub source)
2. **Adaptive Delays**: Dynamic backoff based on rate limit frequency
3. **Circuit Breaker**: Temporary bypass for consistently rate-limited endpoints
4. **Cache Integration**: Reduce API calls through intelligent caching

## Documentation Updates
Updated documentation files:
- `DOCKERHUB_IMPLEMENTATION.md`: Added comprehensive rate limiting section
- `DOCKERHUB_EDGE_CASES_CHECKLIST.md`: Marked rate limiting as fully implemented
- Source code comments: Detailed function documentation

## Conclusion
The implemented rate limiting solution transforms the DockerHub source from a fragile implementation that failed on rate limits to a robust, production-ready scanner that gracefully handles DockerHub's API constraints. This improvement significantly enhances the reliability of DockerHub scans, especially for large organizations or during periods of heavy API usage.