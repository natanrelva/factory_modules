package latency

import (
	"github.com/dubbing-system/audio-interface/pkg/types"
	"testing"
	"time"
)

func TestNewLatencyManager(t *testing.T) {
	target := 100 * time.Millisecond
	lm := NewLatencyManager(target)

	if lm == nil {
		t.Fatal("NewLatencyManager returned nil")
	}

	if lm.GetTargetLatency() != target {
		t.Errorf("Expected target latency %v, got %v", target, lm.GetTargetLatency())
	}

	if lm.GetCurrentMode() != types.Shared {
		t.Errorf("Expected default mode Shared, got %v", lm.GetCurrentMode())
	}
}

func TestLatencyManager_UpdateLatency(t *testing.T) {
	lm := NewLatencyManager(100 * time.Millisecond)

	capture := 25 * time.Millisecond
	playback := 35 * time.Millisecond

	lm.UpdateLatency(capture, playback)

	metrics := lm.MonitorLatency()
	if metrics.CaptureLatency != capture {
		t.Errorf("Expected capture latency %v, got %v", capture, metrics.CaptureLatency)
	}
	if metrics.PlaybackLatency != playback {
		t.Errorf("Expected playback latency %v, got %v", playback, metrics.PlaybackLatency)
	}

	endToEnd := lm.GetEndToEndLatency()
	expected := capture + playback
	if endToEnd != expected {
		t.Errorf("Expected end-to-end latency %v, got %v", expected, endToEnd)
	}
}

func TestLatencyManager_IsWithinTarget(t *testing.T) {
	lm := NewLatencyManager(100 * time.Millisecond)

	// Within target
	lm.UpdateLatency(30*time.Millisecond, 40*time.Millisecond)
	if !lm.IsWithinTarget() {
		t.Error("Expected latency to be within target")
	}

	// Exceeds target
	lm.UpdateLatency(60*time.Millisecond, 50*time.Millisecond)
	if lm.IsWithinTarget() {
		t.Error("Expected latency to exceed target")
	}
}

func TestLatencyManager_SetTargetLatency(t *testing.T) {
	lm := NewLatencyManager(100 * time.Millisecond)

	// Valid target
	err := lm.SetTargetLatency(80 * time.Millisecond)
	if err != nil {
		t.Errorf("SetTargetLatency failed: %v", err)
	}

	if lm.GetTargetLatency() != 80*time.Millisecond {
		t.Errorf("Expected target 80ms, got %v", lm.GetTargetLatency())
	}

	// Invalid targets
	err = lm.SetTargetLatency(5 * time.Millisecond)
	if err == nil {
		t.Error("Expected error for too low target")
	}

	err = lm.SetTargetLatency(600 * time.Millisecond)
	if err == nil {
		t.Error("Expected error for too high target")
	}
}

func TestLatencyManager_OptimizeBuffers(t *testing.T) {
	lm := NewLatencyManager(100 * time.Millisecond)

	// High CPU load
	err := lm.OptimizeBuffers(0.85)
	if err != nil {
		t.Errorf("OptimizeBuffers failed: %v", err)
	}

	stats := lm.GetStats()
	if stats.BufferOptimizations != 1 {
		t.Errorf("Expected 1 optimization, got %d", stats.BufferOptimizations)
	}

	// Cooldown should prevent immediate optimization
	err = lm.OptimizeBuffers(0.9)
	if err == nil {
		t.Error("Expected error due to cooldown")
	}

	// Wait for cooldown
	time.Sleep(1100 * time.Millisecond)

	// Low CPU load
	err = lm.OptimizeBuffers(0.2)
	if err != nil {
		t.Errorf("OptimizeBuffers after cooldown failed: %v", err)
	}

	stats = lm.GetStats()
	if stats.BufferOptimizations != 2 {
		t.Errorf("Expected 2 optimizations, got %d", stats.BufferOptimizations)
	}
}

func TestLatencyManager_OptimizeBuffers_InvalidLoad(t *testing.T) {
	lm := NewLatencyManager(100 * time.Millisecond)

	// Invalid CPU loads
	err := lm.OptimizeBuffers(-0.1)
	if err == nil {
		t.Error("Expected error for negative CPU load")
	}

	err = lm.OptimizeBuffers(1.5)
	if err == nil {
		t.Error("Expected error for CPU load > 1.0")
	}
}

func TestLatencyManager_SelectOperationMode(t *testing.T) {
	lm := NewLatencyManager(40 * time.Millisecond)

	// Low target latency, low CPU load -> Exclusive
	lm.UpdateCPULoad(0.3)
	mode, err := lm.SelectOperationMode()
	if err != nil {
		t.Errorf("SelectOperationMode failed: %v", err)
	}
	if mode != types.Exclusive {
		t.Errorf("Expected Exclusive mode, got %v", mode)
	}

	// High CPU load -> Shared
	lm.UpdateCPULoad(0.8)
	mode, err = lm.SelectOperationMode()
	if err != nil {
		t.Errorf("SelectOperationMode failed: %v", err)
	}
	if mode != types.Shared {
		t.Errorf("Expected Shared mode with high CPU load, got %v", mode)
	}

	// High target latency -> Exclusive (if CPU allows)
	lm.SetTargetLatency(100 * time.Millisecond)
	lm.UpdateCPULoad(0.5)
	mode, err = lm.SelectOperationMode()
	if err != nil {
		t.Errorf("SelectOperationMode failed: %v", err)
	}
	if mode != types.Exclusive {
		t.Errorf("Expected Exclusive mode, got %v", mode)
	}
}

func TestLatencyManager_GetAverageLatency(t *testing.T) {
	lm := NewLatencyManager(100 * time.Millisecond)

	// No data
	avg := lm.GetAverageLatency()
	if avg != 0 {
		t.Errorf("Expected 0 average with no data, got %v", avg)
	}

	// Add some measurements
	lm.UpdateLatency(20*time.Millisecond, 30*time.Millisecond) // 50ms
	lm.UpdateLatency(25*time.Millisecond, 35*time.Millisecond) // 60ms
	lm.UpdateLatency(30*time.Millisecond, 40*time.Millisecond) // 70ms

	avg = lm.GetAverageLatency()
	expected := 60 * time.Millisecond // (50+60+70)/3
	if avg != expected {
		t.Errorf("Expected average %v, got %v", expected, avg)
	}
}

func TestLatencyManager_GetLatencyPercentile(t *testing.T) {
	lm := NewLatencyManager(100 * time.Millisecond)

	// Add measurements
	for i := 0; i < 10; i++ {
		latency := time.Duration(10+i*10) * time.Millisecond
		lm.UpdateLatency(latency/2, latency/2)
	}

	// Test percentiles
	p50 := lm.GetLatencyPercentile(50)
	if p50 < 50*time.Millisecond || p50 > 70*time.Millisecond {
		t.Errorf("P50 latency %v outside expected range", p50)
	}

	p95 := lm.GetLatencyPercentile(95)
	if p95 < 90*time.Millisecond || p95 > 100*time.Millisecond {
		t.Errorf("P95 latency %v outside expected range", p95)
	}

	t.Logf("P50: %v, P95: %v", p50, p95)
}

func TestLatencyManager_GetLatencyPercentile_InvalidPercentile(t *testing.T) {
	lm := NewLatencyManager(100 * time.Millisecond)

	lm.UpdateLatency(20*time.Millisecond, 30*time.Millisecond)

	// Invalid percentiles
	p := lm.GetLatencyPercentile(-10)
	if p != 0 {
		t.Errorf("Expected 0 for invalid percentile, got %v", p)
	}

	p = lm.GetLatencyPercentile(150)
	if p != 0 {
		t.Errorf("Expected 0 for invalid percentile, got %v", p)
	}
}

func TestLatencyManager_LatencyViolations(t *testing.T) {
	lm := NewLatencyManager(100 * time.Millisecond)

	// Within target
	lm.UpdateLatency(30*time.Millisecond, 40*time.Millisecond)

	stats := lm.GetStats()
	if stats.LatencyViolations != 0 {
		t.Errorf("Expected 0 violations, got %d", stats.LatencyViolations)
	}

	// Exceed target by more than 150ms
	lm.UpdateLatency(150*time.Millisecond, 120*time.Millisecond) // 270ms total

	stats = lm.GetStats()
	if stats.LatencyViolations != 1 {
		t.Errorf("Expected 1 violation, got %d", stats.LatencyViolations)
	}
}

func TestLatencyManager_UpdateCPULoad(t *testing.T) {
	lm := NewLatencyManager(100 * time.Millisecond)

	// Valid load
	err := lm.UpdateCPULoad(0.5)
	if err != nil {
		t.Errorf("UpdateCPULoad failed: %v", err)
	}

	stats := lm.GetStats()
	if stats.CPULoad != 0.5 {
		t.Errorf("Expected CPU load 0.5, got %f", stats.CPULoad)
	}

	// Invalid loads
	err = lm.UpdateCPULoad(-0.1)
	if err == nil {
		t.Error("Expected error for negative CPU load")
	}

	err = lm.UpdateCPULoad(1.5)
	if err == nil {
		t.Error("Expected error for CPU load > 1.0")
	}
}

func TestLatencyManager_CalculateJitterCompensation(t *testing.T) {
	lm := NewLatencyManager(100 * time.Millisecond)

	// Not enough data
	jitter := lm.CalculateJitterCompensation()
	if jitter != 0 {
		t.Errorf("Expected 0 jitter with insufficient data, got %v", jitter)
	}

	// Add measurements with varying latency
	lm.UpdateLatency(20*time.Millisecond, 30*time.Millisecond)
	lm.UpdateLatency(25*time.Millisecond, 35*time.Millisecond)
	lm.UpdateLatency(30*time.Millisecond, 40*time.Millisecond)
	lm.UpdateLatency(22*time.Millisecond, 32*time.Millisecond)
	lm.UpdateLatency(28*time.Millisecond, 38*time.Millisecond)

	jitter = lm.CalculateJitterCompensation()
	// Should have some jitter compensation
	t.Logf("Calculated jitter compensation: %v", jitter)

	stats := lm.GetStats()
	if stats.JitterCompensation != jitter {
		t.Errorf("Stats jitter %v doesn't match calculated %v", stats.JitterCompensation, jitter)
	}
}

func TestLatencyManager_Reset(t *testing.T) {
	lm := NewLatencyManager(100 * time.Millisecond)

	// Add some data
	lm.UpdateLatency(20*time.Millisecond, 30*time.Millisecond)
	lm.UpdateLatency(25*time.Millisecond, 35*time.Millisecond)
	lm.OptimizeBuffers(0.9)

	stats := lm.GetStats()
	if stats.EndToEndLatency == 0 {
		t.Error("Expected non-zero latency before reset")
	}

	// Reset
	lm.Reset()

	stats = lm.GetStats()
	if stats.EndToEndLatency != 0 {
		t.Errorf("Expected 0 end-to-end latency after reset, got %v", stats.EndToEndLatency)
	}
	if stats.BufferOptimizations != 0 {
		t.Errorf("Expected 0 optimizations after reset, got %d", stats.BufferOptimizations)
	}
	if stats.LatencyViolations != 0 {
		t.Errorf("Expected 0 violations after reset, got %d", stats.LatencyViolations)
	}
}

func TestLatencyManager_GetStats(t *testing.T) {
	lm := NewLatencyManager(100 * time.Millisecond)

	lm.UpdateLatency(25*time.Millisecond, 35*time.Millisecond)
	lm.UpdateCPULoad(0.6)
	lm.SelectOperationMode()

	stats := lm.GetStats()

	if stats.CaptureLatency != 25*time.Millisecond {
		t.Errorf("Expected capture latency 25ms, got %v", stats.CaptureLatency)
	}
	if stats.PlaybackLatency != 35*time.Millisecond {
		t.Errorf("Expected playback latency 35ms, got %v", stats.PlaybackLatency)
	}
	if stats.EndToEndLatency != 60*time.Millisecond {
		t.Errorf("Expected end-to-end latency 60ms, got %v", stats.EndToEndLatency)
	}
	if stats.TargetLatency != 100*time.Millisecond {
		t.Errorf("Expected target latency 100ms, got %v", stats.TargetLatency)
	}
	if stats.CPULoad != 0.6 {
		t.Errorf("Expected CPU load 0.6, got %f", stats.CPULoad)
	}
}

func TestLatencyManager_HistoryLimit(t *testing.T) {
	lm := NewLatencyManager(100 * time.Millisecond)

	// Add more than maxHistorySize measurements
	for i := 0; i < 150; i++ {
		lm.UpdateLatency(20*time.Millisecond, 30*time.Millisecond)
	}

	// History should be limited
	avg := lm.GetAverageLatency()
	if avg == 0 {
		t.Error("Expected non-zero average")
	}

	// Should still work correctly
	p50 := lm.GetLatencyPercentile(50)
	if p50 == 0 {
		t.Error("Expected non-zero P50")
	}
}

func TestLatencyManager_ConcurrentAccess(t *testing.T) {
	lm := NewLatencyManager(100 * time.Millisecond)

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			lm.UpdateLatency(20*time.Millisecond, 30*time.Millisecond)
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Reader goroutines
	for j := 0; j < 3; j++ {
		go func() {
			for i := 0; i < 100; i++ {
				lm.MonitorLatency()
				lm.GetEndToEndLatency()
				lm.GetAverageLatency()
				lm.GetStats()
				time.Sleep(time.Millisecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 4; i++ {
		<-done
	}

	// No crashes = success
}
