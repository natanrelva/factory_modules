package metrics

import (
	"fmt"
	"github.com/dubbing-system/audio-interface/pkg/types"
	"testing"
	"time"
)

func TestNewMetricsCollector(t *testing.T) {
	mc := NewMetricsCollector()
	if mc == nil {
		t.Fatal("NewMetricsCollector returned nil")
	}

	uptime := mc.GetUptime()
	if uptime < 0 {
		t.Errorf("Invalid uptime: %v", uptime)
	}
}

func TestMetricsCollector_RecordLatency(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordLatency("capture", 25*time.Millisecond)
	mc.RecordLatency("capture", 30*time.Millisecond)
	mc.RecordLatency("playback", 35*time.Millisecond)

	// Check capture metrics
	metrics, err := mc.GetModuleMetrics("capture")
	if err != nil {
		t.Fatalf("GetModuleMetrics failed: %v", err)
	}

	if metrics.LatencyCount != 2 {
		t.Errorf("Expected 2 measurements, got %d", metrics.LatencyCount)
	}

	avgLatency := mc.GetAverageLatency("capture")
	expected := 27500 * time.Microsecond // (25+30)/2 = 27.5ms
	if avgLatency != expected {
		t.Errorf("Expected average latency %v, got %v", expected, avgLatency)
	}

	if metrics.MinLatency != 25*time.Millisecond {
		t.Errorf("Expected min latency 25ms, got %v", metrics.MinLatency)
	}
	if metrics.MaxLatency != 30*time.Millisecond {
		t.Errorf("Expected max latency 30ms, got %v", metrics.MaxLatency)
	}
}

func TestMetricsCollector_RecordError(t *testing.T) {
	mc := NewMetricsCollector()

	errInfo := types.ErrorInfo{
		Module:    "capture",
		Operation: "Initialize",
		Err:       fmt.Errorf("test error"),
		Context:   "test context",
	}

	mc.RecordError(errInfo)

	errorCount := mc.GetErrorCount("capture")
	if errorCount != 1 {
		t.Errorf("Expected 1 error, got %d", errorCount)
	}

	recentErrors := mc.GetRecentErrors(10)
	if len(recentErrors) != 1 {
		t.Errorf("Expected 1 recent error, got %d", len(recentErrors))
	}

	if recentErrors[0].Module != "capture" {
		t.Errorf("Expected module 'capture', got '%s'", recentErrors[0].Module)
	}
}

func TestMetricsCollector_GetModuleMetrics(t *testing.T) {
	mc := NewMetricsCollector()

	// Non-existent module
	_, err := mc.GetModuleMetrics("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent module")
	}

	// Add metrics
	mc.RecordLatency("test", 10*time.Millisecond)

	metrics, err := mc.GetModuleMetrics("test")
	if err != nil {
		t.Errorf("GetModuleMetrics failed: %v", err)
	}

	if metrics.ModuleName != "test" {
		t.Errorf("Expected module name 'test', got '%s'", metrics.ModuleName)
	}
}

func TestMetricsCollector_GetAllModuleMetrics(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordLatency("capture", 20*time.Millisecond)
	mc.RecordLatency("playback", 30*time.Millisecond)
	mc.RecordLatency("sync", 5*time.Millisecond)

	allMetrics := mc.GetAllModuleMetrics()

	if len(allMetrics) != 3 {
		t.Errorf("Expected 3 modules, got %d", len(allMetrics))
	}

	if _, exists := allMetrics["capture"]; !exists {
		t.Error("Expected 'capture' module in metrics")
	}
	if _, exists := allMetrics["playback"]; !exists {
		t.Error("Expected 'playback' module in metrics")
	}
	if _, exists := allMetrics["sync"]; !exists {
		t.Error("Expected 'sync' module in metrics")
	}
}

func TestMetricsCollector_GetAverageLatency(t *testing.T) {
	mc := NewMetricsCollector()

	// No data
	avg := mc.GetAverageLatency("test")
	if avg != 0 {
		t.Errorf("Expected 0 average with no data, got %v", avg)
	}

	// Add measurements
	mc.RecordLatency("test", 10*time.Millisecond)
	mc.RecordLatency("test", 20*time.Millisecond)
	mc.RecordLatency("test", 30*time.Millisecond)

	avg = mc.GetAverageLatency("test")
	expected := 20 * time.Millisecond
	if avg != expected {
		t.Errorf("Expected average %v, got %v", expected, avg)
	}
}

func TestMetricsCollector_GetErrorCount(t *testing.T) {
	mc := NewMetricsCollector()

	// No errors
	count := mc.GetErrorCount("test")
	if count != 0 {
		t.Errorf("Expected 0 errors, got %d", count)
	}

	// Add errors
	for i := 0; i < 5; i++ {
		mc.RecordError(types.ErrorInfo{
			Module:    "test",
			Operation: "TestOp",
		})
	}

	count = mc.GetErrorCount("test")
	if count != 5 {
		t.Errorf("Expected 5 errors, got %d", count)
	}
}

func TestMetricsCollector_GetRecentErrors(t *testing.T) {
	mc := NewMetricsCollector()

	// Add errors
	for i := 0; i < 15; i++ {
		mc.RecordError(types.ErrorInfo{
			Module:    "test",
			Operation: fmt.Sprintf("Op%d", i),
		})
	}

	// Get last 10
	recent := mc.GetRecentErrors(10)
	if len(recent) != 10 {
		t.Errorf("Expected 10 recent errors, got %d", len(recent))
	}

	// Get all
	all := mc.GetRecentErrors(0)
	if len(all) != 15 {
		t.Errorf("Expected 15 errors, got %d", len(all))
	}

	// Get more than available
	many := mc.GetRecentErrors(100)
	if len(many) != 15 {
		t.Errorf("Expected 15 errors, got %d", len(many))
	}
}

func TestMetricsCollector_GetErrorsByModule(t *testing.T) {
	mc := NewMetricsCollector()

	// Add errors for different modules
	mc.RecordError(types.ErrorInfo{Module: "capture", Operation: "Op1"})
	mc.RecordError(types.ErrorInfo{Module: "playback", Operation: "Op2"})
	mc.RecordError(types.ErrorInfo{Module: "capture", Operation: "Op3"})

	captureErrors := mc.GetErrorsByModule("capture")
	if len(captureErrors) != 2 {
		t.Errorf("Expected 2 capture errors, got %d", len(captureErrors))
	}

	playbackErrors := mc.GetErrorsByModule("playback")
	if len(playbackErrors) != 1 {
		t.Errorf("Expected 1 playback error, got %d", len(playbackErrors))
	}

	nonexistentErrors := mc.GetErrorsByModule("nonexistent")
	if len(nonexistentErrors) != 0 {
		t.Errorf("Expected 0 errors for nonexistent module, got %d", len(nonexistentErrors))
	}
}

func TestMetricsCollector_Reset(t *testing.T) {
	mc := NewMetricsCollector()

	// Add data
	mc.RecordLatency("test", 10*time.Millisecond)
	mc.RecordError(types.ErrorInfo{Module: "test", Operation: "TestOp"})

	// Verify data exists
	if len(mc.GetAllModuleMetrics()) == 0 {
		t.Error("Expected metrics before reset")
	}

	// Reset
	mc.Reset()

	// Verify data cleared
	if len(mc.GetAllModuleMetrics()) != 0 {
		t.Error("Expected no metrics after reset")
	}

	recentErrors := mc.GetRecentErrors(10)
	if len(recentErrors) != 0 {
		t.Errorf("Expected no errors after reset, got %d", len(recentErrors))
	}
}

func TestMetricsCollector_GetUptime(t *testing.T) {
	mc := NewMetricsCollector()

	time.Sleep(100 * time.Millisecond)

	uptime := mc.GetUptime()
	if uptime < 100*time.Millisecond {
		t.Errorf("Expected uptime >= 100ms, got %v", uptime)
	}
	if uptime > 200*time.Millisecond {
		t.Errorf("Expected uptime < 200ms, got %v", uptime)
	}
}

func TestMetricsCollector_GetSummary(t *testing.T) {
	mc := NewMetricsCollector()

	// Add metrics
	mc.RecordLatency("capture", 20*time.Millisecond)
	mc.RecordLatency("capture", 30*time.Millisecond)
	mc.RecordLatency("playback", 40*time.Millisecond)
	mc.RecordError(types.ErrorInfo{Module: "capture", Operation: "TestOp"})

	summary := mc.GetSummary()

	if summary.TotalModules != 2 {
		t.Errorf("Expected 2 modules, got %d", summary.TotalModules)
	}

	if summary.TotalErrors != 1 {
		t.Errorf("Expected 1 error, got %d", summary.TotalErrors)
	}

	if summary.Uptime < 0 {
		t.Errorf("Expected non-negative uptime, got %v", summary.Uptime)
	}

	captureSummary, exists := summary.ModuleSummary["capture"]
	if !exists {
		t.Fatal("Expected capture module in summary")
	}

	if captureSummary.MeasurementCount != 2 {
		t.Errorf("Expected 2 measurements, got %d", captureSummary.MeasurementCount)
	}

	if captureSummary.ErrorCount != 1 {
		t.Errorf("Expected 1 error, got %d", captureSummary.ErrorCount)
	}
}

func TestMetricsCollector_ExportMetrics(t *testing.T) {
	mc := NewMetricsCollector()

	// Add data
	mc.RecordLatency("capture", 20*time.Millisecond)
	mc.RecordLatency("playback", 30*time.Millisecond)
	mc.RecordError(types.ErrorInfo{Module: "test", Operation: "TestOp"})

	export := mc.ExportMetrics()

	if export.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}

	if export.Uptime < 0 {
		t.Errorf("Expected non-negative uptime, got %v", export.Uptime)
	}

	// Note: RecordError also creates a module entry
	if len(export.Modules) < 2 {
		t.Errorf("Expected at least 2 modules, got %d", len(export.Modules))
	}

	if len(export.RecentErrors) != 1 {
		t.Errorf("Expected 1 recent error, got %d", len(export.RecentErrors))
	}
}

func TestMetricsCollector_ErrorHistoryLimit(t *testing.T) {
	mc := NewMetricsCollector()

	// Add more than maxErrorHistory errors
	for i := 0; i < 150; i++ {
		mc.RecordError(types.ErrorInfo{
			Module:    "test",
			Operation: fmt.Sprintf("Op%d", i),
		})
	}

	recentErrors := mc.GetRecentErrors(0)
	if len(recentErrors) > 100 {
		t.Errorf("Expected max 100 errors in history, got %d", len(recentErrors))
	}
}

func TestMetricsCollector_ConcurrentAccess(t *testing.T) {
	mc := NewMetricsCollector()

	done := make(chan bool)

	// Writer goroutines
	for i := 0; i < 3; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				mc.RecordLatency(fmt.Sprintf("module%d", id), time.Duration(j)*time.Millisecond)
				mc.RecordError(types.ErrorInfo{
					Module:    fmt.Sprintf("module%d", id),
					Operation: "TestOp",
				})
				time.Sleep(time.Millisecond)
			}
			done <- true
		}(i)
	}

	// Reader goroutines
	for i := 0; i < 2; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				mc.GetAllModuleMetrics()
				mc.GetRecentErrors(10)
				mc.GetSummary()
				time.Sleep(time.Millisecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// No crashes = success
	summary := mc.GetSummary()
	if summary.TotalModules == 0 {
		t.Error("Expected some modules after concurrent access")
	}
}
