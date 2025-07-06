package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/trufflesecurity/trufflehog/v3/pkg/engine"
	"github.com/trufflesecurity/trufflehog/v3/pkg/output"
	"github.com/trufflesecurity/trufflehog/v3/pkg/sources"
	"github.com/trufflesecurity/trufflehog/v3/pkg/sources/github"
)

// BenchmarkResult represents the performance analysis results
type BenchmarkResult struct {
	TotalScanDuration         time.Duration              `json:"total_scan_duration"`
	TotalBytesScanned         uint64                     `json:"total_bytes_scanned"`
	TotalChunksScanned        uint64                     `json:"total_chunks_scanned"`
	TotalSecretsFound         uint64                     `json:"total_secrets_found"`
	DecoderPerformance        map[string]DecoderMetrics  `json:"decoder_performance"`
	AhoCorasickPerformance    AhoCorasickMetrics         `json:"ahocorasick_performance"`
	DetectorPerformance       map[string]DetectorMetrics `json:"detector_performance"`
	VerificationPerformance   map[string]VerificationMetrics `json:"verification_performance"`
	MemoryUsage               MemoryMetrics              `json:"memory_usage"`
	BottleneckAnalysis        BottleneckAnalysis         `json:"bottleneck_analysis"`
}

type DecoderMetrics struct {
	TotalExecutionTime time.Duration `json:"total_execution_time"`
	TotalInputBytes    uint64        `json:"total_input_bytes"`
	TotalOutputBytes   uint64        `json:"total_output_bytes"`
	SuccessCount       uint64        `json:"success_count"`
	AvgExecutionTime   time.Duration `json:"avg_execution_time"`
	Throughput         float64       `json:"throughput_bytes_per_second"`
}

type AhoCorasickMetrics struct {
	TotalExecutionTime   time.Duration `json:"total_execution_time"`
	TotalInputBytes      uint64        `json:"total_input_bytes"`
	TotalKeywordMatches  uint64        `json:"total_keyword_matches"`
	TotalDetectorMatches uint64        `json:"total_detector_matches"`
	AvgExecutionTime     time.Duration `json:"avg_execution_time"`
	Throughput           float64       `json:"throughput_bytes_per_second"`
}

type DetectorMetrics struct {
	TotalRegexTime       time.Duration `json:"total_regex_time"`
	TotalVerificationTime time.Duration `json:"total_verification_time"`
	TotalInputBytes      uint64        `json:"total_input_bytes"`
	TotalMatches         uint64        `json:"total_matches"`
	AvgRegexTime         time.Duration `json:"avg_regex_time"`
	RegexThroughput      float64       `json:"regex_throughput_bytes_per_second"`
}

type VerificationMetrics struct {
	TotalAttempts       uint64        `json:"total_attempts"`
	SuccessfulAttempts  uint64        `json:"successful_attempts"`
	TotalExecutionTime  time.Duration `json:"total_execution_time"`
	AvgExecutionTime    time.Duration `json:"avg_execution_time"`
	SuccessRate         float64       `json:"success_rate"`
	ErrorsByType        map[string]uint64 `json:"errors_by_type"`
}

type MemoryMetrics struct {
	PeakHeapAlloc     uint64 `json:"peak_heap_alloc"`
	PeakHeapSys       uint64 `json:"peak_heap_sys"`
	PeakGoroutines    uint64 `json:"peak_goroutines"`
	AvgChannelQueueSize map[string]float64 `json:"avg_channel_queue_size"`
}

type BottleneckAnalysis struct {
	SlowestStage        string        `json:"slowest_stage"`
	SlowestStageTime    time.Duration `json:"slowest_stage_time"`
	SlowestDetector     string        `json:"slowest_detector"`
	SlowestDetectorTime time.Duration `json:"slowest_detector_time"`
	Recommendations     []string      `json:"recommendations"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run benchmark_figma.go <github_token>")
	}

	githubToken := os.Args[1]

	// Start Prometheus metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":2112", nil))
	}()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupts
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nReceived interrupt, shutting down...")
		cancel()
	}()

	// Run benchmark
	result, err := runBenchmark(ctx, githubToken)
	if err != nil {
		log.Fatalf("Benchmark failed: %v", err)
	}

	// Output results
	outputResults(result)
}

func runBenchmark(ctx context.Context, githubToken string) (*BenchmarkResult, error) {
	fmt.Println("🚀 Starting TruffleHog Figma Organization Benchmark")
	fmt.Printf("🔍 Scanning ~170 repositories in the Figma organization\n")
	fmt.Printf("📊 Prometheus metrics available at http://localhost:2112/metrics\n\n")

	// Create GitHub source
	githubSource := github.Source{}
	githubConn, err := githubSource.Init(ctx, "trufflehog", 0, 0, false,
		&sources.Config{}, sources.WithConcurrency(runtime.NumCPU()))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GitHub source: %w", err)
	}

	// Create source manager
	sourceManager := sources.NewManager(
		sources.WithConcurrency(runtime.NumCPU()),
		sources.WithBufferSize(10000),
	)

	// Create TruffleHog engine
	engineConfig := &engine.Config{
		Concurrency:        runtime.NumCPU(),
		Verify:             true,
		Dispatcher:         engine.NewPrinterDispatcher(new(output.JSONPrinter)),
		SourceManager:      sourceManager,
		PrintAvgDetectorTime: true,
	}

	eng, err := engine.NewEngine(ctx, engineConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create engine: %w", err)
	}

	// Start metrics collection
	metricsCollector := NewMetricsCollector()
	go metricsCollector.Start(ctx)

	// Configure GitHub source to scan Figma organization
	githubConfig := &github.Config{
		Token:        githubToken,
		Endpoint:     "https://api.github.com",
		Orgs:         []string{"figma"},
		Concurrency:  runtime.NumCPU(),
		IncludeForks: false,
	}

	// Start scanning
	fmt.Println("📈 Starting scan...")
	startTime := time.Now()

	eng.Start(ctx)
	
	// Add GitHub source to scan Figma organization
	err = sourceManager.Run(ctx, "github-figma", githubSource.Init(ctx, "trufflehog", 0, 0, false, &sources.Config{}, sources.WithConcurrency(runtime.NumCPU())))
	if err != nil {
		return nil, fmt.Errorf("failed to run GitHub source: %w", err)
	}

	// Wait for completion or cancellation
	select {
	case <-ctx.Done():
		fmt.Println("\n⏹️  Scan cancelled")
	case <-time.After(time.Hour * 2): // 2 hour timeout
		fmt.Println("\n⏱️  Scan timed out after 2 hours")
		cancel()
	}

	// Finish engine
	err = eng.Finish(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to finish engine: %w", err)
	}

	scanDuration := time.Since(startTime)
	fmt.Printf("✅ Scan completed in %v\n", scanDuration)

	// Collect and analyze metrics
	result := analyzeMetrics(eng, metricsCollector, scanDuration)
	
	return result, nil
}

func analyzeMetrics(eng *engine.Engine, collector *MetricsCollector, scanDuration time.Duration) *BenchmarkResult {
	fmt.Println("\n📊 Analyzing performance metrics...")

	metrics := eng.GetMetrics()
	
	result := &BenchmarkResult{
		TotalScanDuration:  scanDuration,
		TotalBytesScanned:  metrics.BytesScanned,
		TotalChunksScanned: metrics.ChunksScanned,
		TotalSecretsFound:  metrics.VerifiedSecretsFound + metrics.UnverifiedSecretsFound,
		DecoderPerformance: make(map[string]DecoderMetrics),
		DetectorPerformance: make(map[string]DetectorMetrics),
		VerificationPerformance: make(map[string]VerificationMetrics),
		MemoryUsage:        collector.GetMemoryMetrics(),
	}

	// Analyze decoder performance
	result.DecoderPerformance = collector.GetDecoderMetrics()

	// Analyze Aho-Corasick performance
	result.AhoCorasickPerformance = collector.GetAhoCorasickMetrics()

	// Analyze detector performance
	result.DetectorPerformance = collector.GetDetectorMetrics()

	// Analyze verification performance
	result.VerificationPerformance = collector.GetVerificationMetrics()

	// Perform bottleneck analysis
	result.BottleneckAnalysis = analyzeBottlenecks(result)

	return result
}

func analyzeBottlenecks(result *BenchmarkResult) BottleneckAnalysis {
	analysis := BottleneckAnalysis{
		Recommendations: []string{},
	}

	// Find slowest pipeline stage
	stages := map[string]time.Duration{
		"decoding":       getTotalDecoderTime(result.DecoderPerformance),
		"ahocorasick":    result.AhoCorasickPerformance.TotalExecutionTime,
		"regex":          getTotalRegexTime(result.DetectorPerformance),
		"verification":   getTotalVerificationTime(result.VerificationPerformance),
	}

	var slowestStage string
	var slowestTime time.Duration
	for stage, duration := range stages {
		if duration > slowestTime {
			slowestTime = duration
			slowestStage = stage
		}
	}

	analysis.SlowestStage = slowestStage
	analysis.SlowestStageTime = slowestTime

	// Find slowest detector
	var slowestDetector string
	var slowestDetectorTime time.Duration
	for detector, metrics := range result.DetectorPerformance {
		totalTime := metrics.TotalRegexTime + metrics.TotalVerificationTime
		if totalTime > slowestDetectorTime {
			slowestDetectorTime = totalTime
			slowestDetector = detector
		}
	}

	analysis.SlowestDetector = slowestDetector
	analysis.SlowestDetectorTime = slowestDetectorTime

	// Generate recommendations
	analysis.Recommendations = generateRecommendations(result, analysis)

	return analysis
}

func getTotalDecoderTime(decoders map[string]DecoderMetrics) time.Duration {
	var total time.Duration
	for _, metrics := range decoders {
		total += metrics.TotalExecutionTime
	}
	return total
}

func getTotalRegexTime(detectors map[string]DetectorMetrics) time.Duration {
	var total time.Duration
	for _, metrics := range detectors {
		total += metrics.TotalRegexTime
	}
	return total
}

func getTotalVerificationTime(verifications map[string]VerificationMetrics) time.Duration {
	var total time.Duration
	for _, metrics := range verifications {
		total += metrics.TotalExecutionTime
	}
	return total
}

func generateRecommendations(result *BenchmarkResult, analysis BottleneckAnalysis) []string {
	recommendations := []string{}

	// Performance recommendations based on bottleneck analysis
	switch analysis.SlowestStage {
	case "decoding":
		recommendations = append(recommendations, 
			"🔧 Decoding is the bottleneck. Consider optimizing decoder implementations or reducing decoder variety.",
			"💡 Consider parallel decoding or async processing for large chunks.")
	case "ahocorasick":
		recommendations = append(recommendations,
			"🔧 Aho-Corasick keyword matching is the bottleneck. Consider optimizing keyword lists or using more efficient data structures.",
			"💡 Consider reducing the number of keywords or improving the Aho-Corasick implementation.")
	case "regex":
		recommendations = append(recommendations,
			"🔧 Regex matching is the bottleneck. Focus on optimizing regex patterns and detector implementations.",
			"💡 Consider using more efficient regex engines or pre-compiled patterns.")
	case "verification":
		recommendations = append(recommendations,
			"🔧 Verification network calls are the bottleneck. Consider optimizing network requests or reducing verification frequency.",
			"💡 Implement request batching, connection pooling, or async verification.")
	}

	// Memory recommendations
	if result.MemoryUsage.PeakHeapAlloc > 1<<30 { // 1GB
		recommendations = append(recommendations,
			"💾 High memory usage detected. Consider implementing memory pooling or reducing buffer sizes.")
	}

	// Concurrency recommendations
	if result.MemoryUsage.PeakGoroutines > 1000 {
		recommendations = append(recommendations,
			"🔄 High goroutine count detected. Consider reducing concurrency or implementing worker pools.")
	}

	return recommendations
}

func outputResults(result *BenchmarkResult) {
	fmt.Println("\n" + "="*80)
	fmt.Println("📈 TRUFFLEHOG FIGMA ORGANIZATION BENCHMARK RESULTS")
	fmt.Println("="*80)

	fmt.Printf("⏱️  Total Scan Duration: %v\n", result.TotalScanDuration)
	fmt.Printf("📊 Total Bytes Scanned: %d (%.2f MB)\n", result.TotalBytesScanned, float64(result.TotalBytesScanned)/1024/1024)
	fmt.Printf("🔍 Total Chunks Scanned: %d\n", result.TotalChunksScanned)
	fmt.Printf("🔑 Total Secrets Found: %d\n", result.TotalSecretsFound)
	fmt.Printf("🚀 Throughput: %.2f MB/s\n", float64(result.TotalBytesScanned)/result.TotalScanDuration.Seconds()/1024/1024)

	fmt.Println("\n🎯 BOTTLENECK ANALYSIS")
	fmt.Println("-"*40)
	fmt.Printf("🐌 Slowest Stage: %s (%.2f seconds)\n", result.BottleneckAnalysis.SlowestStage, result.BottleneckAnalysis.SlowestStageTime.Seconds())
	fmt.Printf("🔍 Slowest Detector: %s (%.2f seconds)\n", result.BottleneckAnalysis.SlowestDetector, result.BottleneckAnalysis.SlowestDetectorTime.Seconds())

	fmt.Println("\n📋 RECOMMENDATIONS")
	fmt.Println("-"*40)
	for _, rec := range result.BottleneckAnalysis.Recommendations {
		fmt.Printf("  %s\n", rec)
	}

	fmt.Println("\n🔧 DETAILED PERFORMANCE BREAKDOWN")
	fmt.Println("-"*40)
	
	// Decoder performance
	fmt.Println("📝 Decoder Performance:")
	for decoder, metrics := range result.DecoderPerformance {
		fmt.Printf("  %s: %.2fs (%.2f MB/s)\n", decoder, metrics.TotalExecutionTime.Seconds(), metrics.Throughput/1024/1024)
	}

	// Aho-Corasick performance
	fmt.Printf("\n🔍 Aho-Corasick Performance: %.2fs (%.2f MB/s)\n", 
		result.AhoCorasickPerformance.TotalExecutionTime.Seconds(), 
		result.AhoCorasickPerformance.Throughput/1024/1024)

	// Top 10 slowest detectors
	fmt.Println("\n🔍 Top 10 Slowest Detectors:")
	type detectorTime struct {
		name string
		time time.Duration
	}
	
	var detectors []detectorTime
	for name, metrics := range result.DetectorPerformance {
		totalTime := metrics.TotalRegexTime + metrics.TotalVerificationTime
		detectors = append(detectors, detectorTime{name, totalTime})
	}
	
	sort.Slice(detectors, func(i, j int) bool {
		return detectors[i].time > detectors[j].time
	})
	
	for i, detector := range detectors {
		if i >= 10 {
			break
		}
		fmt.Printf("  %d. %s: %.2fs\n", i+1, detector.name, detector.time.Seconds())
	}

	// Memory usage
	fmt.Println("\n💾 Memory Usage:")
	fmt.Printf("  Peak Heap Allocation: %.2f MB\n", float64(result.MemoryUsage.PeakHeapAlloc)/1024/1024)
	fmt.Printf("  Peak Heap System: %.2f MB\n", float64(result.MemoryUsage.PeakHeapSys)/1024/1024)
	fmt.Printf("  Peak Goroutines: %d\n", result.MemoryUsage.PeakGoroutines)

	// Save detailed results to JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err == nil {
		filename := fmt.Sprintf("trufflehog_benchmark_%s.json", time.Now().Format("2006-01-02_15-04-05"))
		err = os.WriteFile(filename, jsonData, 0644)
		if err == nil {
			fmt.Printf("\n💾 Detailed results saved to: %s\n", filename)
		}
	}

	fmt.Println("\n✅ Benchmark completed successfully!")
}

// MetricsCollector helps collect and aggregate Prometheus metrics
type MetricsCollector struct {
	// Implementation would collect metrics from Prometheus
	// For now, we'll simulate with basic data structures
}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

func (mc *MetricsCollector) Start(ctx context.Context) {
	// Implementation would start collecting metrics
}

func (mc *MetricsCollector) GetDecoderMetrics() map[string]DecoderMetrics {
	// Implementation would return actual decoder metrics
	return make(map[string]DecoderMetrics)
}

func (mc *MetricsCollector) GetAhoCorasickMetrics() AhoCorasickMetrics {
	// Implementation would return actual Aho-Corasick metrics
	return AhoCorasickMetrics{}
}

func (mc *MetricsCollector) GetDetectorMetrics() map[string]DetectorMetrics {
	// Implementation would return actual detector metrics
	return make(map[string]DetectorMetrics)
}

func (mc *MetricsCollector) GetVerificationMetrics() map[string]VerificationMetrics {
	// Implementation would return actual verification metrics
	return make(map[string]VerificationMetrics)
}

func (mc *MetricsCollector) GetMemoryMetrics() MemoryMetrics {
	// Implementation would return actual memory metrics
	return MemoryMetrics{}
}