package metrics

import (
	"fmt"
	"github.com/dubbing-system/audio-interface/pkg/types"
	"sync"
	"time"
)

// MetricsCollector collects and aggregates metrics from all audio modules
type MetricsCollector struct {
	mu              sync.RWMutex
	moduleMetrics   map[string]*ModuleMetrics
	errors          []types.ErrorInfo
	maxErrorHistory int
	startTime       time.Time
}

// ModuleMetrics contains metrics for a specific module
type ModuleMetrics struct {
	ModuleName       string
	LatencySum       time.Duration
	LatencyCount     int64
	MinLatency       time.Duration
	MaxLatency       time.Duration
	ErrorCount       int64
	LastUpdateTime   time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		moduleMetrics:   make(map[string]*ModuleMetrics),
		errors:          make([]types.ErrorInfo, 0),
		maxErrorHistory: 100,
		startTime:       time.Now(),
	}
}

// RecordLatency records a latency measurement for a module
func (mc *MetricsCollector) RecordLatency(module string, latency time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	metrics, exists := mc.moduleMetrics[module]
	if !exists {
		metrics = &ModuleMetrics{
			ModuleName:  module,
			MinLatency:  latency,
			MaxLatency:  latency,
		}
		mc.moduleMetrics[module] = metrics
	}

	// Update metrics
	metrics.LatencySum += latency
	metrics.LatencyCount++
	metrics.LastUpdateTime = time.Now()

	if latency < metrics.MinLatency {
		metrics.MinLatency = latency
	}
	if latency > metrics.MaxLatency {
		metrics.MaxLatency = latency
	}
}

// RecordError records an error for a module
func (mc *MetricsCollector) RecordError(errInfo types.ErrorInfo) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Add timestamp if not set
	if errInfo.Timestamp.IsZero() {
		errInfo.Timestamp = time.Now()
	}

	// Add to error history
	mc.errors = append(mc.errors, errInfo)

	// Trim history if too large
	if len(mc.errors) > mc.maxErrorHistory {
		mc.errors = mc.errors[1:]
	}

	// Update module error count
	metrics, exists := mc.moduleMetrics[errInfo.Module]
	if !exists {
		metrics = &ModuleMetrics{
			ModuleName: errInfo.Module,
		}
		mc.moduleMetrics[errInfo.Module] = metrics
	}
	metrics.ErrorCount++
}

// GetMetrics returns current latency metrics (implements interfaces.MetricsCollector)
func (mc *MetricsCollector) GetMetrics() types.LatencyMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Aggregate metrics from all modules
	var totalLatency time.Duration
	var totalCount int64

	for _, metrics := range mc.moduleMetrics {
		totalLatency += metrics.LatencySum
		totalCount += metrics.LatencyCount
	}

	avgLatency := time.Duration(0)
	if totalCount > 0 {
		avgLatency = totalLatency / time.Duration(totalCount)
	}

	return types.LatencyMetrics{
		CaptureLatency:  avgLatency,
		PlaybackLatency: avgLatency,
		BufferFillLevel: 0.0,
		DroppedFrames:   0,
		Underruns:       0,
		Overruns:        0,
		Timestamp:       time.Now(),
	}
}

// GetModuleMetrics returns metrics for a specific module
func (mc *MetricsCollector) GetModuleMetrics(module string) (*ModuleMetrics, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics, exists := mc.moduleMetrics[module]
	if !exists {
		return nil, fmt.Errorf("no metrics found for module: %s", module)
	}

	// Return a copy
	metricsCopy := *metrics
	return &metricsCopy, nil
}

// GetAllModuleMetrics returns metrics for all modules
func (mc *MetricsCollector) GetAllModuleMetrics() map[string]*ModuleMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Return copies
	result := make(map[string]*ModuleMetrics)
	for name, metrics := range mc.moduleMetrics {
		metricsCopy := *metrics
		result[name] = &metricsCopy
	}

	return result
}

// GetAverageLatency returns average latency for a module
func (mc *MetricsCollector) GetAverageLatency(module string) time.Duration {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics, exists := mc.moduleMetrics[module]
	if !exists || metrics.LatencyCount == 0 {
		return 0
	}

	return metrics.LatencySum / time.Duration(metrics.LatencyCount)
}

// GetErrorCount returns total error count for a module
func (mc *MetricsCollector) GetErrorCount(module string) int64 {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics, exists := mc.moduleMetrics[module]
	if !exists {
		return 0
	}

	return metrics.ErrorCount
}

// GetRecentErrors returns recent errors (up to maxCount)
func (mc *MetricsCollector) GetRecentErrors(maxCount int) []types.ErrorInfo {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if maxCount <= 0 || maxCount > len(mc.errors) {
		maxCount = len(mc.errors)
	}

	// Return most recent errors
	start := len(mc.errors) - maxCount
	result := make([]types.ErrorInfo, maxCount)
	copy(result, mc.errors[start:])

	return result
}

// GetErrorsByModule returns errors for a specific module
func (mc *MetricsCollector) GetErrorsByModule(module string) []types.ErrorInfo {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make([]types.ErrorInfo, 0)
	for _, err := range mc.errors {
		if err.Module == module {
			result = append(result, err)
		}
	}

	return result
}

// Reset clears all metrics and error history
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.moduleMetrics = make(map[string]*ModuleMetrics)
	mc.errors = make([]types.ErrorInfo, 0)
	mc.startTime = time.Now()
}

// GetUptime returns the time since metrics collection started
func (mc *MetricsCollector) GetUptime() time.Duration {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return time.Since(mc.startTime)
}

// GetSummary returns a summary of all metrics
func (mc *MetricsCollector) GetSummary() MetricsSummary {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	summary := MetricsSummary{
		Uptime:        time.Since(mc.startTime),
		TotalModules:  len(mc.moduleMetrics),
		TotalErrors:   len(mc.errors),
		ModuleSummary: make(map[string]ModuleSummary),
	}

	for name, metrics := range mc.moduleMetrics {
		avgLatency := time.Duration(0)
		if metrics.LatencyCount > 0 {
			avgLatency = metrics.LatencySum / time.Duration(metrics.LatencyCount)
		}

		summary.ModuleSummary[name] = ModuleSummary{
			ModuleName:      name,
			AverageLatency:  avgLatency,
			MinLatency:      metrics.MinLatency,
			MaxLatency:      metrics.MaxLatency,
			MeasurementCount: metrics.LatencyCount,
			ErrorCount:      metrics.ErrorCount,
			LastUpdate:      metrics.LastUpdateTime,
		}
	}

	return summary
}

// ExportMetrics exports metrics in a structured format
func (mc *MetricsCollector) ExportMetrics() MetricsExport {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	export := MetricsExport{
		Timestamp:     time.Now(),
		Uptime:        time.Since(mc.startTime),
		Modules:       make([]ModuleMetrics, 0, len(mc.moduleMetrics)),
		RecentErrors:  make([]types.ErrorInfo, 0),
	}

	// Export module metrics
	for _, metrics := range mc.moduleMetrics {
		metricsCopy := *metrics
		export.Modules = append(export.Modules, metricsCopy)
	}

	// Export recent errors (last 10)
	errorCount := len(mc.errors)
	if errorCount > 10 {
		errorCount = 10
	}
	if errorCount > 0 {
		start := len(mc.errors) - errorCount
		export.RecentErrors = append(export.RecentErrors, mc.errors[start:]...)
	}

	return export
}

// MetricsSummary provides a high-level summary of metrics
type MetricsSummary struct {
	Uptime        time.Duration
	TotalModules  int
	TotalErrors   int
	ModuleSummary map[string]ModuleSummary
}

// ModuleSummary provides summary for a single module
type ModuleSummary struct {
	ModuleName       string
	AverageLatency   time.Duration
	MinLatency       time.Duration
	MaxLatency       time.Duration
	MeasurementCount int64
	ErrorCount       int64
	LastUpdate       time.Time
}

// MetricsExport provides exportable metrics data
type MetricsExport struct {
	Timestamp    time.Time
	Uptime       time.Duration
	Modules      []ModuleMetrics
	RecentErrors []types.ErrorInfo
}
