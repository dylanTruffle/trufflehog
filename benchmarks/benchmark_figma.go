package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/model"
)

type BenchmarkResults struct {
	ScanDuration              string                 `json:"scan_duration"`
	TotalRepositories         int                    `json:"total_repositories"`
	OverallThroughputMBps     float64               `json:"overall_throughput_mbps"`
	PipelineBottlenecks       []BottleneckAnalysis  `json:"pipeline_bottlenecks"`
	ConcurrencyMetrics        ConcurrencyAnalysis   `json:"concurrency_metrics"`
	TopSlowDetectors          []DetectorPerformance `json:"top_slow_detectors"`
	SystemResourceUsage       SystemMetrics         `json:"system_resource_usage"`
	Recommendations           []string              `json:"recommendations"`
	PrometheusMetricsURL      string                `json:"prometheus_metrics_url"`
}

type BottleneckAnalysis struct {
	Stage                string  `json:"stage"`
	ChannelUtilization   float64 `json:"channel_utilization_percent"`
	BlockedWrites        float64 `json:"blocked_writes_per_second"`
	AverageLatency       float64 `json:"average_latency_ms"`
	BottleneckSeverity   string  `json:"bottleneck_severity"`
}

type ConcurrencyAnalysis struct {
	WorkerUtilization    map[string]float64 `json:"worker_utilization_percent"`
	WorkerStarvation     map[string]float64 `json:"worker_starvation_events"`
	ChannelSaturation    map[string]float64 `json:"channel_saturation_percent"`
	ConcurrencyEfficiency float64           `json:"concurrency_efficiency_percent"`
}

type DetectorPerformance struct {
	DetectorName         string  `json:"detector_name"`
	AverageLatency       float64 `json:"average_latency_ms"`
	ThroughputPerSecond  float64 `json:"throughput_per_second"`
	VerificationLatency  float64 `json:"verification_latency_ms"`
	ImpactOnPipeline     string  `json:"impact_on_pipeline"`
}

type SystemMetrics struct {
	PeakMemoryUsageMB    float64 `json:"peak_memory_usage_mb"`
	AverageGoroutines    float64 `json:"average_goroutines"`
	GCPauseTimeP95       float64 `json:"gc_pause_time_p95_ms"`
	SystemBottlenecks    []string `json:"system_bottlenecks"`
}

func main() {
	fmt.Println("🔍 TruffleHog Pipeline Performance Benchmark - Figma Organization")
	fmt.Println(strings.Repeat("=", 80))
	
	// Start Prometheus metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Prometheus metrics available at http://localhost:2112/metrics")
		log.Fatal(http.ListenAndServe(":2112", nil))
	}()
	
	// Wait for metrics server to start
	time.Sleep(2 * time.Second)
	
	// Run TruffleHog scan with pipeline-focused flags
	fmt.Println("🚀 Starting TruffleHog scan of Figma organization...")
	startTime := time.Now()
	
	cmd := exec.Command("../trufflehog", 
		"github", 
		"--org=figma",
		"--no-update",  // Disable updates as requested
		"--concurrency=16", // Higher concurrency to stress test pipeline
		"--json",
		"--no-only-verified", // Include unverified results for throughput testing
	)
	
	output, err := cmd.CombinedOutput()
	scanDuration := time.Since(startTime)
	
	if err != nil {
		log.Printf("TruffleHog scan completed with warnings: %v", err)
	}
	
	fmt.Printf("✅ Scan completed in %v\n", scanDuration)
	
	// Parse scan results to count repositories
	repoCount := countRepositories(string(output))
	fmt.Printf("📊 Scanned %d repositories\n", repoCount)
	
	// Wait a moment for final metrics to be recorded
	time.Sleep(3 * time.Second)
	
	// Collect and analyze pipeline metrics
	fmt.Println("\n📈 Analyzing pipeline performance metrics...")
	results, err := analyzePipelineMetrics(scanDuration, repoCount)
	if err != nil {
		log.Fatalf("Failed to analyze metrics: %v", err)
	}
	
	// Display results
	displayResults(results)
	
	// Save detailed results to JSON
	saveResultsToFile(results)
	
	fmt.Printf("\n🎯 Benchmark complete! Prometheus metrics available at %s\n", results.PrometheusMetricsURL)
}

func countRepositories(output string) int {
	lines := strings.Split(output, "\n")
	repoSet := make(map[string]bool)
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(line), &result); err == nil {
			if sourceMetadata, ok := result["SourceMetadata"].(map[string]interface{}); ok {
				if data, ok := sourceMetadata["Data"].(map[string]interface{}); ok {
					if github, ok := data["Github"].(map[string]interface{}); ok {
						if repo, ok := github["repository"].(string); ok {
							repoSet[repo] = true
						}
					}
				}
			}
		}
	}
	
	return len(repoSet)
}

func analyzePipelineMetrics(scanDuration time.Duration, repoCount int) (*BenchmarkResults, error) {
	// Connect to Prometheus
	client, err := api.NewClient(api.Config{
		Address: "http://localhost:2112",
	})
	if err != nil {
		return nil, fmt.Errorf("error creating prometheus client: %w", err)
	}
	
	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	results := &BenchmarkResults{
		ScanDuration:         scanDuration.String(),
		TotalRepositories:    repoCount,
		PrometheusMetricsURL: "http://localhost:2112/metrics",
	}
	
	// Analyze overall throughput
	if throughput, err := queryPrometheusGauge(ctx, v1api, "trufflehog_overall_throughput_bytes_per_second"); err == nil {
		results.OverallThroughputMBps = throughput / (1024 * 1024) // Convert to MB/s
	}
	
	// Analyze pipeline bottlenecks
	results.PipelineBottlenecks = analyzePipelineBottlenecks(ctx, v1api)
	
	// Analyze concurrency metrics
	results.ConcurrencyMetrics = analyzeConcurrencyMetrics(ctx, v1api)
	
	// Analyze detector performance
	results.TopSlowDetectors = analyzeDetectorPerformance(ctx, v1api)
	
	// Analyze system resource usage
	results.SystemResourceUsage = analyzeSystemMetrics(ctx, v1api)
	
	// Generate recommendations
	results.Recommendations = generateRecommendations(results)
	
	return results, nil
}

func analyzePipelineBottlenecks(ctx context.Context, v1api v1.API) []BottleneckAnalysis {
	var bottlenecks []BottleneckAnalysis
	
	channels := []string{"chunks", "detectable_chunks", "verification_overlap", "results"}
	
	for _, channel := range channels {
		analysis := BottleneckAnalysis{Stage: channel}
		
		// Channel utilization
		if depth, err := queryPrometheusGauge(ctx, v1api, fmt.Sprintf("trufflehog_%s_channel_queue_depth", channel)); err == nil {
			// Estimate capacity based on TruffleHog's channel buffer sizes
			capacity := 50.0 * 50.0 // detectableChunksChanMultiplier * defaultChannelBuffer
			if channel == "verification_overlap" {
				capacity = 25.0 * 50.0
			} else if channel == "results" {
				capacity = 50.0 // defaultChannelBuffer
			}
			analysis.ChannelUtilization = (depth / capacity) * 100
		}
		
		// Blocked writes
		if blocked, err := queryPrometheusCounter(ctx, v1api, fmt.Sprintf("trufflehog_channel_blocked_writes_total{channel=\"%s\"}", channel)); err == nil {
			analysis.BlockedWrites = blocked
		}
		
		// Average latency for this stage
		if latency, err := queryPrometheusHistogramAvg(ctx, v1api, fmt.Sprintf("trufflehog_stage_latency_microseconds{stage=\"%s\"}", channel)); err == nil {
			analysis.AverageLatency = latency / 1000 // Convert to milliseconds
		}
		
		// Determine bottleneck severity
		if analysis.ChannelUtilization > 80 || analysis.BlockedWrites > 10 {
			analysis.BottleneckSeverity = "HIGH"
		} else if analysis.ChannelUtilization > 50 || analysis.BlockedWrites > 1 {
			analysis.BottleneckSeverity = "MEDIUM"
		} else {
			analysis.BottleneckSeverity = "LOW"
		}
		
		bottlenecks = append(bottlenecks, analysis)
	}
	
	// Sort by severity
	sort.Slice(bottlenecks, func(i, j int) bool {
		severityOrder := map[string]int{"HIGH": 3, "MEDIUM": 2, "LOW": 1}
		return severityOrder[bottlenecks[i].BottleneckSeverity] > severityOrder[bottlenecks[j].BottleneckSeverity]
	})
	
	return bottlenecks
}

func analyzeConcurrencyMetrics(ctx context.Context, v1api v1.API) ConcurrencyAnalysis {
	analysis := ConcurrencyAnalysis{
		WorkerUtilization: make(map[string]float64),
		WorkerStarvation:  make(map[string]float64),
		ChannelSaturation: make(map[string]float64),
	}
	
	workerTypes := []string{"scanner", "detector", "verification_overlap", "notifier"}
	
	for _, workerType := range workerTypes {
		// Worker utilization
		if active, err := queryPrometheusGauge(ctx, v1api, fmt.Sprintf("trufflehog_active_workers{worker_type=\"%s\"}", workerType)); err == nil {
			// Estimate max workers based on TruffleHog's worker setup
			maxWorkers := 16.0 // concurrency setting
			if workerType == "detector" || workerType == "verification_overlap" {
				maxWorkers = 16.0 * 50.0 // concurrency * detectorWorkerMultiplier
			} else if workerType == "notifier" {
				maxWorkers = 16.0 / 4.0 // concurrency / notifierWorkerRatio
			}
			analysis.WorkerUtilization[workerType] = (active / maxWorkers) * 100
		}
		
		// Worker starvation
		if starvation, err := queryPrometheusCounter(ctx, v1api, fmt.Sprintf("trufflehog_worker_starvation_total{worker_type=\"%s\"}", workerType)); err == nil {
			analysis.WorkerStarvation[workerType] = starvation
		}
	}
	
	// Calculate overall concurrency efficiency
	totalUtilization := 0.0
	for _, util := range analysis.WorkerUtilization {
		totalUtilization += util
	}
	if len(analysis.WorkerUtilization) > 0 {
		analysis.ConcurrencyEfficiency = totalUtilization / float64(len(analysis.WorkerUtilization))
	}
	
	return analysis
}

func analyzeDetectorPerformance(ctx context.Context, v1api v1.API) []DetectorPerformance {
	var detectors []DetectorPerformance
	
	// Get list of detectors from metrics
	detectorNames, err := getDetectorNames(ctx, v1api)
	if err != nil {
		return detectors
	}
	
	for _, detector := range detectorNames {
		perf := DetectorPerformance{DetectorName: detector}
		
		// Average execution latency
		if latency, err := queryPrometheusHistogramAvg(ctx, v1api, fmt.Sprintf("trufflehog_detector_execution_time_microseconds{detector=\"%s\"}", detector)); err == nil {
			perf.AverageLatency = latency / 1000 // Convert to milliseconds
		}
		
		// Throughput
		if throughput, err := queryPrometheusCounter(ctx, v1api, fmt.Sprintf("trufflehog_detector_chunks_processed_total{detector=\"%s\"}", detector)); err == nil {
			perf.ThroughputPerSecond = throughput
		}
		
		// Verification latency
		if verifyLatency, err := queryPrometheusHistogramAvg(ctx, v1api, fmt.Sprintf("trufflehog_verification_latency_milliseconds{detector=\"%s\"}", detector)); err == nil {
			perf.VerificationLatency = verifyLatency
		}
		
		// Determine impact on pipeline
		if perf.AverageLatency > 1000 { // > 1 second
			perf.ImpactOnPipeline = "HIGH"
		} else if perf.AverageLatency > 100 { // > 100ms
			perf.ImpactOnPipeline = "MEDIUM"
		} else {
			perf.ImpactOnPipeline = "LOW"
		}
		
		detectors = append(detectors, perf)
	}
	
	// Sort by average latency (descending)
	sort.Slice(detectors, func(i, j int) bool {
		return detectors[i].AverageLatency > detectors[j].AverageLatency
	})
	
	// Return top 10 slowest detectors
	if len(detectors) > 10 {
		detectors = detectors[:10]
	}
	
	return detectors
}

func analyzeSystemMetrics(ctx context.Context, v1api v1.API) SystemMetrics {
	metrics := SystemMetrics{}
	
	// Peak memory usage
	if memory, err := queryPrometheusGauge(ctx, v1api, "trufflehog_memory_usage_bytes"); err == nil {
		metrics.PeakMemoryUsageMB = memory / (1024 * 1024)
	}
	
	// Average goroutines
	if goroutines, err := queryPrometheusGauge(ctx, v1api, "trufflehog_goroutines_total"); err == nil {
		metrics.AverageGoroutines = goroutines
	}
	
	// GC pause time P95
	if gcPause, err := queryPrometheusHistogramQuantile(ctx, v1api, "trufflehog_gc_pause_time_microseconds", 0.95); err == nil {
		metrics.GCPauseTimeP95 = gcPause / 1000 // Convert to milliseconds
	}
	
	// Identify system bottlenecks
	if metrics.PeakMemoryUsageMB > 8192 { // > 8GB
		metrics.SystemBottlenecks = append(metrics.SystemBottlenecks, "High memory usage detected")
	}
	if metrics.AverageGoroutines > 10000 {
		metrics.SystemBottlenecks = append(metrics.SystemBottlenecks, "High goroutine count detected")
	}
	if metrics.GCPauseTimeP95 > 100 { // > 100ms
		metrics.SystemBottlenecks = append(metrics.SystemBottlenecks, "High GC pause times detected")
	}
	
	return metrics
}

func generateRecommendations(results *BenchmarkResults) []string {
	var recommendations []string
	
	// Pipeline bottleneck recommendations
	for _, bottleneck := range results.PipelineBottlenecks {
		if bottleneck.BottleneckSeverity == "HIGH" {
			switch bottleneck.Stage {
			case "detectable_chunks":
				recommendations = append(recommendations, 
					"🔧 CRITICAL: Detector workers are saturated. Consider increasing detector worker multiplier or reducing detector count.")
			case "verification_overlap":
				recommendations = append(recommendations, 
					"🔧 CRITICAL: Verification overlap processing is bottlenecked. Consider disabling verification overlap with --allow-verification-overlap.")
			case "results":
				recommendations = append(recommendations, 
					"🔧 CRITICAL: Results notification is slow. Consider optimizing output format or increasing notifier workers.")
			case "chunks":
				recommendations = append(recommendations, 
					"🔧 CRITICAL: Input chunk processing is bottlenecked. Consider reducing concurrency or optimizing source reading.")
			}
		}
	}
	
	// Concurrency efficiency recommendations
	if results.ConcurrencyMetrics.ConcurrencyEfficiency < 50 {
		recommendations = append(recommendations, 
			"⚡ Worker efficiency is low. Consider reducing concurrency to match available work or optimizing worker algorithms.")
	}
	
	// Detector performance recommendations
	for _, detector := range results.TopSlowDetectors {
		if detector.ImpactOnPipeline == "HIGH" && detector.AverageLatency > 1000 {
			recommendations = append(recommendations, 
				fmt.Sprintf("🐌 Detector '%s' is very slow (%.1fms avg). Consider excluding with --exclude-detectors=%s", 
					detector.DetectorName, detector.AverageLatency, detector.DetectorName))
		}
	}
	
	// System resource recommendations
	if results.SystemResourceUsage.PeakMemoryUsageMB > 8192 {
		recommendations = append(recommendations, 
			"💾 High memory usage detected. Consider reducing concurrency or implementing memory optimizations.")
	}
	
	// Throughput recommendations
	if results.OverallThroughputMBps < 10 { // Less than 10 MB/s
		recommendations = append(recommendations, 
			"🚀 Low overall throughput. Focus on the highest severity bottlenecks identified above.")
	}
	
	// General optimization recommendations
	recommendations = append(recommendations, 
		"📊 Monitor Prometheus metrics at "+results.PrometheusMetricsURL+" for real-time pipeline analysis")
	recommendations = append(recommendations, 
		"🎯 For production scans, focus on optimizing the bottlenecks with 'HIGH' severity first")
	
	return recommendations
}

// Prometheus query helper functions
func queryPrometheusGauge(ctx context.Context, v1api v1.API, query string) (float64, error) {
	result, _, err := v1api.Query(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}
	
	if vector, ok := result.(model.Vector); ok && len(vector) > 0 {
		return float64(vector[0].Value), nil
	}
	
	return 0, fmt.Errorf("no data found for query: %s", query)
}

func queryPrometheusCounter(ctx context.Context, v1api v1.API, query string) (float64, error) {
	return queryPrometheusGauge(ctx, v1api, fmt.Sprintf("rate(%s[5m])", query))
}

func queryPrometheusHistogramAvg(ctx context.Context, v1api v1.API, metric string) (float64, error) {
	query := fmt.Sprintf("rate(%s_sum[5m]) / rate(%s_count[5m])", metric, metric)
	return queryPrometheusGauge(ctx, v1api, query)
}

func queryPrometheusHistogramQuantile(ctx context.Context, v1api v1.API, metric string, quantile float64) (float64, error) {
	query := fmt.Sprintf("histogram_quantile(%.2f, rate(%s_bucket[5m]))", quantile, metric)
	return queryPrometheusGauge(ctx, v1api, query)
}

func getDetectorNames(ctx context.Context, v1api v1.API) ([]string, error) {
	result, _, err := v1api.Query(ctx, "group by (detector) (trufflehog_detector_execution_time_microseconds)", time.Now())
	if err != nil {
		return nil, err
	}
	
	var detectors []string
	if vector, ok := result.(model.Vector); ok {
		for _, sample := range vector {
			if detector, exists := sample.Metric["detector"]; exists {
				detectors = append(detectors, string(detector))
			}
		}
	}
	
	return detectors, nil
}

func displayResults(results *BenchmarkResults) {
	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("🎯 TRUFFLEHOG PIPELINE PERFORMANCE ANALYSIS\n")
	fmt.Printf(strings.Repeat("=", 80) + "\n")
	
	fmt.Printf("📊 SCAN OVERVIEW:\n")
	fmt.Printf("   Duration: %s\n", results.ScanDuration)
	fmt.Printf("   Repositories: %d\n", results.TotalRepositories)
	fmt.Printf("   Throughput: %.2f MB/s\n", results.OverallThroughputMBps)
	
	fmt.Printf("\n🚨 PIPELINE BOTTLENECKS (by severity):\n")
	for i, bottleneck := range results.PipelineBottlenecks {
		severity := "🟢"
		if bottleneck.BottleneckSeverity == "HIGH" {
			severity = "🔴"
		} else if bottleneck.BottleneckSeverity == "MEDIUM" {
			severity = "🟡"
		}
		
		fmt.Printf("   %d. %s %s: %.1f%% utilization, %.1f blocked writes/sec, %.1fms avg latency\n",
			i+1, severity, bottleneck.Stage, bottleneck.ChannelUtilization, 
			bottleneck.BlockedWrites, bottleneck.AverageLatency)
	}
	
	fmt.Printf("\n⚡ CONCURRENCY ANALYSIS:\n")
	fmt.Printf("   Overall Efficiency: %.1f%%\n", results.ConcurrencyMetrics.ConcurrencyEfficiency)
	fmt.Printf("   Worker Utilization:\n")
	for worker, util := range results.ConcurrencyMetrics.WorkerUtilization {
		fmt.Printf("     %s: %.1f%%\n", worker, util)
	}
	
	fmt.Printf("\n🐌 SLOWEST DETECTORS:\n")
	for i, detector := range results.TopSlowDetectors {
		if i >= 5 { // Show top 5
			break
		}
		impact := "🟢"
		if detector.ImpactOnPipeline == "HIGH" {
			impact = "🔴"
		} else if detector.ImpactOnPipeline == "MEDIUM" {
			impact = "🟡"
		}
		
		fmt.Printf("   %d. %s %s: %.1fms avg (%.1f/sec throughput)\n",
			i+1, impact, detector.DetectorName, detector.AverageLatency, detector.ThroughputPerSecond)
	}
	
	fmt.Printf("\n💾 SYSTEM RESOURCES:\n")
	fmt.Printf("   Peak Memory: %.1f MB\n", results.SystemResourceUsage.PeakMemoryUsageMB)
	fmt.Printf("   Avg Goroutines: %.0f\n", results.SystemResourceUsage.AverageGoroutines)
	fmt.Printf("   GC Pause P95: %.1fms\n", results.SystemResourceUsage.GCPauseTimeP95)
	
	fmt.Printf("\n🎯 RECOMMENDATIONS:\n")
	for i, rec := range results.Recommendations {
		fmt.Printf("   %d. %s\n", i+1, rec)
	}
}

func saveResultsToFile(results *BenchmarkResults) {
	filename := fmt.Sprintf("trufflehog_pipeline_benchmark_%s.json", 
		time.Now().Format("2006-01-02_15-04-05"))
	
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to create results file: %v", err)
		return
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		log.Printf("Failed to write results: %v", err)
		return
	}
	
	fmt.Printf("\n💾 Detailed results saved to: %s\n", filename)
}