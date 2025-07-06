package engine

import (
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/trufflesecurity/trufflehog/v3/pkg/common"
)

var (
	decodeLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "decode_latency",
			Help:      "Time spent decoding a chunk in microseconds",
			Buckets:   prometheus.ExponentialBuckets(50, 2, 20),
		},
		[]string{"decoder_type", "source_name"},
	)

	// Detector metrics.
	detectorExecutionCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "detector_execution_count",
			Help:      "Total number of times a detector has been executed.",
		},
		[]string{"detector_name", "job_id", "source_name"},
	)

	// Note this is the time taken to call FromData on each detector, not necessarily the time taken
	// to verify a credential via an API call. If the regex match within FromData does not match, the
	// detector will return early. For now this is a good proxy for the time taken to verify a credential.
	// TODO (ahrav)
	// We can work on a more fine-grained metric later.
	detectorExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "detector_execution_duration",
			Help:      "Duration of detector execution in milliseconds.",
			Buckets:   prometheus.ExponentialBuckets(1, 5, 6),
		},
		[]string{"detector_name"},
	)

	jobBytesScanned = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: common.MetricsNamespace,
		Subsystem: common.MetricsSubsystem,
		Name:      "job_bytes_scanned",
		Help:      "Total number of bytes scanned for a job.",
	},
		[]string{"source_type", "source_name"},
	)

	scanBytesPerChunk = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: common.MetricsNamespace,
		Subsystem: common.MetricsSubsystem,
		Name:      "scan_bytes_per_chunk",
		Help:      "Total number of bytes in a chunk.",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 18),
	})

	jobChunksScanned = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: common.MetricsNamespace,
		Subsystem: common.MetricsSubsystem,
		Name:      "job_chunks_scanned",
		Help:      "Total number of chunks scanned for a job.",
	},
		[]string{"source_type", "source_name"},
	)

	detectBytesPerMatch = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: common.MetricsNamespace,
		Subsystem: common.MetricsSubsystem,
		Name:      "detect_bytes_per_match",
		Help:      "Total number of bytes used to detect a credential in a match per chunk.",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 18),
	})

	matchesPerChunk = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: common.MetricsNamespace,
		Subsystem: common.MetricsSubsystem,
		Name:      "matches_per_chunk",
		Help:      "Total number of matches found in a chunk.",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 10),
	})

	// Metrics around latency for the different stages of the pipeline.
	chunksScannedLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: common.MetricsNamespace,
		Subsystem: common.MetricsSubsystem,
		Name:      "chunk_scanned_latency",
		Help:      "Time taken to scan a chunk in microseconds.",
		Buckets:   prometheus.ExponentialBuckets(1, 2, 22),
	})

	chunksDetectedLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: common.MetricsNamespace,
		Subsystem: common.MetricsSubsystem,
		Name:      "chunk_detected_latency",
		Help:      "Time taken to detect a chunk in microseconds.",
		Buckets:   prometheus.ExponentialBuckets(50, 2, 20),
	})

	chunksNotifiedLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: common.MetricsNamespace,
		Subsystem: common.MetricsSubsystem,
		Name:      "chunk_notified_latency",
		Help:      "Time taken to notify a chunk in milliseconds.",
		Buckets:   prometheus.ExponentialBuckets(5, 2, 12),
	})

	// NEW GRANULAR PERFORMANCE METRICS FOR BOTTLENECK ANALYSIS
	
	// Decoder stage metrics
	decoderExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "decoder_execution_duration",
			Help:      "Duration of decoder execution in microseconds.",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 18),
		},
		[]string{"decoder_type"},
	)

	decoderInputBytesSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "decoder_input_bytes_size",
			Help:      "Size of input data sent to decoder in bytes.",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 18),
		},
		[]string{"decoder_type"},
	)

	decoderOutputBytesSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "decoder_output_bytes_size",
			Help:      "Size of output data from decoder in bytes.",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 18),
		},
		[]string{"decoder_type"},
	)

	decoderSuccessCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "decoder_success_count",
			Help:      "Number of successful decoder executions.",
		},
		[]string{"decoder_type"},
	)

	// Aho-Corasick stage metrics
	ahocorasickExecutionDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "ahocorasick_execution_duration",
			Help:      "Duration of Aho-Corasick keyword matching in microseconds.",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 18),
		},
	)

	ahocorasickInputBytesSize = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "ahocorasick_input_bytes_size",
			Help:      "Size of input data sent to Aho-Corasick in bytes.",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 18),
		},
	)

	ahocorasickKeywordMatches = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "ahocorasick_keyword_matches",
			Help:      "Number of keyword matches found by Aho-Corasick.",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 10),
		},
	)

	ahocorasickDetectorMatches = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "ahocorasick_detector_matches",
			Help:      "Number of detector matches found by Aho-Corasick.",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 10),
		},
	)

	// Regex stage metrics (separate from verification)
	regexExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "regex_execution_duration",
			Help:      "Duration of regex pattern matching in microseconds.",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 18),
		},
		[]string{"detector_name"},
	)

	regexInputBytesSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "regex_input_bytes_size",
			Help:      "Size of input data sent to regex matching in bytes.",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 18),
		},
		[]string{"detector_name"},
	)

	regexMatchesFound = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "regex_matches_found",
			Help:      "Number of regex matches found per detector call.",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 10),
		},
		[]string{"detector_name"},
	)

	// Verification stage metrics (network calls)
	verificationExecutionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "verification_execution_duration",
			Help:      "Duration of verification network calls in milliseconds.",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 15),
		},
		[]string{"detector_name"},
	)

	verificationAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "verification_attempts",
			Help:      "Total number of verification attempts.",
		},
		[]string{"detector_name"},
	)

	verificationSuccess = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "verification_success",
			Help:      "Total number of successful verifications.",
		},
		[]string{"detector_name"},
	)

	verificationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "verification_errors",
			Help:      "Total number of verification errors.",
		},
		[]string{"detector_name", "error_type"},
	)

	// Memory usage metrics
	systemMemoryUsage = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: common.MetricsNamespace,
		Subsystem: common.MetricsSubsystem,
		Name:      "memory_usage_bytes",
		Help:      "Current memory usage in bytes.",
	})

	// Pipeline stage breakdown
	pipelineStageLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "pipeline_stage_latency",
			Help:      "Time taken for each pipeline stage in microseconds.",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 18),
		},
		[]string{"stage"},
	)

	// Concurrent processing metrics
	concurrentWorkers = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "concurrent_workers",
			Help:      "Number of concurrent workers active.",
		},
		[]string{"worker_type"},
	)

	channelQueueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "channel_queue_size",
			Help:      "Current size of processing channels.",
		},
		[]string{"channel_type"},
	)
)

// Pipeline Throughput Metrics - these show the overall flow rate and bottlenecks
var (
	// Channel Queue Depth - Critical for identifying pipeline bottlenecks
	chunksChannelQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "trufflehog_chunks_channel_queue_depth",
		Help: "Current number of chunks waiting in the input channel",
	})
	
	detectableChunksChannelQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "trufflehog_detectable_chunks_channel_queue_depth", 
		Help: "Current number of chunks waiting for detector processing",
	})
	
	verificationOverlapChannelQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "trufflehog_verification_overlap_channel_queue_depth",
		Help: "Current number of chunks waiting for verification overlap processing",
	})
	
	resultsChannelQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "trufflehog_results_channel_queue_depth",
		Help: "Current number of results waiting for notification",
	})

	// Worker Pool Utilization - Shows if we're saturating workers
	activeWorkers = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "trufflehog_active_workers",
		Help: "Number of workers currently processing",
	}, []string{"worker_type"})
	
	workerWaitTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "trufflehog_worker_wait_time_microseconds",
		Help: "Time workers spend waiting for work",
		Buckets: prometheus.ExponentialBuckets(100, 2, 16), // 100μs to ~6.5s
	}, []string{"worker_type"})

	// Pipeline Stage Throughput - Items per second through each stage
	pipelineThroughput = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trufflehog_pipeline_items_processed_total",
		Help: "Total items processed by each pipeline stage",
	}, []string{"stage"})
	
	// End-to-End Pipeline Latency - Time from chunk input to final result
	pipelineEndToEndLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "trufflehog_pipeline_end_to_end_latency_milliseconds",
		Help: "Time from chunk input to result output",
		Buckets: prometheus.ExponentialBuckets(1, 2, 20), // 1ms to ~17 minutes
	})

	// Stage-specific latencies (when they matter for bottleneck analysis)
	stageLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "trufflehog_stage_latency_microseconds", 
		Help: "Processing time for each pipeline stage",
		Buckets: prometheus.ExponentialBuckets(10, 2, 20), // 10μs to ~10s
	}, []string{"stage", "detector"})

	// Concurrency Saturation Indicators
	channelBlockedWrites = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trufflehog_channel_blocked_writes_total",
		Help: "Number of times workers blocked writing to channels",
	}, []string{"channel"})
	
	workerStarvation = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trufflehog_worker_starvation_total",
		Help: "Number of times workers had no work available",
	}, []string{"worker_type"})

	// Resource Utilization
	goroutineCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "trufflehog_goroutines_total",
		Help: "Current number of goroutines",
	})
	
	gcPauseTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "trufflehog_gc_pause_time_microseconds",
		Help: "Garbage collection pause times",
		Buckets: prometheus.ExponentialBuckets(10, 2, 16),
	})

	// Detector-specific metrics (only when they impact overall throughput)
	detectorExecutionTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "trufflehog_detector_execution_time_microseconds",
		Help: "Time spent in detector execution (regex + verification combined)",
		Buckets: prometheus.ExponentialBuckets(100, 2, 16), // 100μs to ~6.5s
	}, []string{"detector", "verification_enabled"})
	
	detectorThroughput = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "trufflehog_detector_chunks_processed_total",
		Help: "Total chunks processed by each detector",
	}, []string{"detector"})

	// Network verification bottlenecks
	verificationLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "trufflehog_verification_latency_milliseconds",
		Help: "Network verification latency",
		Buckets: prometheus.ExponentialBuckets(10, 2, 16), // 10ms to ~10 minutes
	}, []string{"detector", "result"})
	
	verificationConcurrency = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "trufflehog_active_verifications",
		Help: "Number of active network verifications",
	}, []string{"detector"})

	// Scan Progress and Overall Health
	scanProgress = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "trufflehog_scan_progress_ratio",
		Help: "Scan progress as a ratio (0.0 to 1.0)",
	}, []string{"source"})
	
	overallThroughputBytesPerSecond = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "trufflehog_overall_throughput_bytes_per_second",
		Help: "Overall scanning throughput in bytes per second",
	})
)

// PipelineMetrics tracks the flow through the scanning pipeline
type PipelineMetrics struct {
	// Channel monitoring
	chunksChannelSize                 int64
	detectableChunksChannelSize       int64
	verificationOverlapChannelSize    int64
	resultsChannelSize                int64
	
	// Worker pool states
	activeScannerWorkers              int64
	activeDetectorWorkers             int64
	activeVerificationOverlapWorkers  int64
	activeNotifierWorkers             int64
	
	// Throughput tracking
	chunksProcessedPerSecond          uint64
	bytesProcessedPerSecond           uint64
	resultsGeneratedPerSecond         uint64
	
	// Timing for end-to-end analysis
	scanStartTime                     time.Time
	lastThroughputUpdate              time.Time
	totalChunksProcessed              uint64
	totalBytesProcessed               uint64
	totalResultsGenerated             uint64
}

// Global pipeline metrics instance
var pipelineMetrics = &PipelineMetrics{
	scanStartTime: time.Now(),
	lastThroughputUpdate: time.Now(),
}

// Channel monitoring functions
func RecordChannelQueueDepth(channelName string, depth int) {
	switch channelName {
	case "chunks":
		chunksChannelQueueDepth.Set(float64(depth))
		atomic.StoreInt64(&pipelineMetrics.chunksChannelSize, int64(depth))
	case "detectable_chunks":
		detectableChunksChannelQueueDepth.Set(float64(depth))
		atomic.StoreInt64(&pipelineMetrics.detectableChunksChannelSize, int64(depth))
	case "verification_overlap":
		verificationOverlapChannelQueueDepth.Set(float64(depth))
		atomic.StoreInt64(&pipelineMetrics.verificationOverlapChannelSize, int64(depth))
	case "results":
		resultsChannelQueueDepth.Set(float64(depth))
		atomic.StoreInt64(&pipelineMetrics.resultsChannelSize, int64(depth))
	}
}

// Worker activity tracking
func RecordWorkerActivity(workerType string, isActive bool) {
	var delta float64 = -1
	if isActive {
		delta = 1
	}
	activeWorkers.WithLabelValues(workerType).Add(delta)
	
	// Update pipeline metrics
	switch workerType {
	case "scanner":
		if isActive {
			atomic.AddInt64(&pipelineMetrics.activeScannerWorkers, 1)
		} else {
			atomic.AddInt64(&pipelineMetrics.activeScannerWorkers, -1)
		}
	case "detector":
		if isActive {
			atomic.AddInt64(&pipelineMetrics.activeDetectorWorkers, 1)
		} else {
			atomic.AddInt64(&pipelineMetrics.activeDetectorWorkers, -1)
		}
	case "verification_overlap":
		if isActive {
			atomic.AddInt64(&pipelineMetrics.activeVerificationOverlapWorkers, 1)
		} else {
			atomic.AddInt64(&pipelineMetrics.activeVerificationOverlapWorkers, -1)
		}
	case "notifier":
		if isActive {
			atomic.AddInt64(&pipelineMetrics.activeNotifierWorkers, 1)
		} else {
			atomic.AddInt64(&pipelineMetrics.activeNotifierWorkers, -1)
		}
	}
}

// Record when workers wait for work (indicates potential saturation)
func RecordWorkerWait(workerType string, waitTime time.Duration) {
	workerWaitTime.WithLabelValues(workerType).Observe(float64(waitTime.Microseconds()))
	if waitTime > time.Millisecond {
		workerStarvation.WithLabelValues(workerType).Inc()
	}
}

// Record pipeline stage processing
func RecordPipelineStage(stage string, processingTime time.Duration) {
	pipelineThroughput.WithLabelValues(stage).Inc()
	stageLatency.WithLabelValues(stage, "").Observe(float64(processingTime.Microseconds()))
}

// Record end-to-end pipeline latency
func RecordPipelineEndToEnd(latency time.Duration) {
	pipelineEndToEndLatency.Observe(float64(latency.Milliseconds()))
}

// Record channel blocking (critical bottleneck indicator)
func RecordChannelBlocked(channelName string) {
	channelBlockedWrites.WithLabelValues(channelName).Inc()
}

// Record detector execution (combined regex + verification)
func RecordDetectorExecution(detector string, verificationEnabled bool, duration time.Duration) {
	verification := "false"
	if verificationEnabled {
		verification = "true"
	}
	detectorExecutionTime.WithLabelValues(detector, verification).Observe(float64(duration.Microseconds()))
	detectorThroughput.WithLabelValues(detector).Inc()
}

// Record network verification specifically
func RecordVerification(detector string, duration time.Duration, success bool) {
	result := "success"
	if !success {
		result = "failure"
	}
	verificationLatency.WithLabelValues(detector, result).Observe(float64(duration.Milliseconds()))
}

// Track active verifications for concurrency monitoring
func RecordActiveVerification(detector string, isStarting bool) {
	var delta float64 = -1
	if isStarting {
		delta = 1
	}
	verificationConcurrency.WithLabelValues(detector).Add(delta)
}

// Update throughput calculations
func UpdateThroughputMetrics() {
	now := time.Now()
	elapsed := now.Sub(pipelineMetrics.lastThroughputUpdate)
	if elapsed < time.Second {
		return // Update at most once per second
	}
	
	totalElapsed := now.Sub(pipelineMetrics.scanStartTime)
	if totalElapsed < time.Second {
		return
	}
	
	totalChunks := atomic.LoadUint64(&pipelineMetrics.totalChunksProcessed)
	totalBytes := atomic.LoadUint64(&pipelineMetrics.totalBytesProcessed)
	
	chunksPerSecond := float64(totalChunks) / totalElapsed.Seconds()
	bytesPerSecond := float64(totalBytes) / totalElapsed.Seconds()
	
	// Store as uint64 bit patterns
	atomic.StoreUint64(&pipelineMetrics.chunksProcessedPerSecond, *(*uint64)(unsafe.Pointer(&chunksPerSecond)))
	atomic.StoreUint64(&pipelineMetrics.bytesProcessedPerSecond, *(*uint64)(unsafe.Pointer(&bytesPerSecond)))
	
	overallThroughputBytesPerSecond.Set(bytesPerSecond)
	pipelineMetrics.lastThroughputUpdate = now
}

// Record system metrics
func RecordSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	systemMemoryUsage.Set(float64(m.Alloc))
	goroutineCount.Set(float64(runtime.NumGoroutine()))
	
	// Record GC pause time
	if len(m.PauseNs) > 0 {
		lastPause := m.PauseNs[(m.NumGC+255)%256]
		gcPauseTime.Observe(float64(lastPause / 1000)) // Convert to microseconds
	}
}

// Increment processed counters
func IncrementChunksProcessed(bytes uint64) {
	atomic.AddUint64(&pipelineMetrics.totalChunksProcessed, 1)
	atomic.AddUint64(&pipelineMetrics.totalBytesProcessed, bytes)
}

func IncrementResultsGenerated() {
	atomic.AddUint64(&pipelineMetrics.totalResultsGenerated, 1)
}

// Get current pipeline state for analysis
func GetPipelineState() *PipelineMetrics {
	return &PipelineMetrics{
		chunksChannelSize:                atomic.LoadInt64(&pipelineMetrics.chunksChannelSize),
		detectableChunksChannelSize:      atomic.LoadInt64(&pipelineMetrics.detectableChunksChannelSize),
		verificationOverlapChannelSize:   atomic.LoadInt64(&pipelineMetrics.verificationOverlapChannelSize),
		resultsChannelSize:               atomic.LoadInt64(&pipelineMetrics.resultsChannelSize),
		activeScannerWorkers:             atomic.LoadInt64(&pipelineMetrics.activeScannerWorkers),
		activeDetectorWorkers:            atomic.LoadInt64(&pipelineMetrics.activeDetectorWorkers),
		activeVerificationOverlapWorkers: atomic.LoadInt64(&pipelineMetrics.activeVerificationOverlapWorkers),
		activeNotifierWorkers:            atomic.LoadInt64(&pipelineMetrics.activeNotifierWorkers),
		chunksProcessedPerSecond:         *(*uint64)(unsafe.Pointer(&pipelineMetrics.chunksProcessedPerSecond)),
		bytesProcessedPerSecond:          *(*uint64)(unsafe.Pointer(&pipelineMetrics.bytesProcessedPerSecond)),
		totalChunksProcessed:             atomic.LoadUint64(&pipelineMetrics.totalChunksProcessed),
		totalBytesProcessed:              atomic.LoadUint64(&pipelineMetrics.totalBytesProcessed),
		totalResultsGenerated:            atomic.LoadUint64(&pipelineMetrics.totalResultsGenerated),
		scanStartTime:                    pipelineMetrics.scanStartTime,
	}
}
