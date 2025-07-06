package engine

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/trufflesecurity/trufflehog/v3/pkg/common"
)

var (
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
		[]string{"job_id", "source_type", "source_name"},
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
		[]string{"job_id", "source_type", "source_name"},
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
	memoryUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: common.MetricsNamespace,
			Subsystem: common.MetricsSubsystem,
			Name:      "memory_usage_bytes",
			Help:      "Current memory usage in bytes.",
		},
		[]string{"stage"},
	)

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
