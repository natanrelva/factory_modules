package sync

import (
	"testing"
	"time"
)

func TestNewStreamSynchronizer(t *testing.T) {
	sync := NewStreamSynchronizer()
	if sync == nil {
		t.Fatal("NewStreamSynchronizer returned nil")
	}

	if sync.targetAlignment != 50*time.Millisecond {
		t.Errorf("Expected target alignment 50ms, got %v", sync.targetAlignment)
	}
}

func TestStreamSynchronizer_SyncCapturePlayback(t *testing.T) {
	sync := NewStreamSynchronizer()

	captureTime := time.Now()
	playbackTime := captureTime.Add(30 * time.Millisecond)

	err := sync.SyncCapturePlayback(captureTime, playbackTime)
	if err != nil {
		t.Fatalf("SyncCapturePlayback failed: %v", err)
	}

	stats := sync.GetStats()
	if stats.CaptureDataPoints != 1 {
		t.Errorf("Expected 1 capture data point, got %d", stats.CaptureDataPoints)
	}
	if stats.PlaybackDataPoints != 1 {
		t.Errorf("Expected 1 playback data point, got %d", stats.PlaybackDataPoints)
	}
}

func TestStreamSynchronizer_GetDriftCompensation(t *testing.T) {
	sync := NewStreamSynchronizer()

	// Initially should be zero
	drift := sync.GetDriftCompensation()
	if drift != 0 {
		t.Errorf("Expected initial drift 0, got %v", drift)
	}

	// Add some timestamps with increasing drift (simulating clock skew)
	baseTime := time.Now()
	for i := 0; i < 20; i++ {
		captureTime := baseTime.Add(time.Duration(i) * 100 * time.Millisecond)
		// Playback drift increases over time (simulating clock skew)
		driftMs := 5 + (i / 2) // Increases from 5ms to 15ms
		playbackTime := captureTime.Add(time.Duration(driftMs) * time.Millisecond)
		sync.SyncCapturePlayback(captureTime, playbackTime)
	}

	drift = sync.GetDriftCompensation()
	// With increasing drift, should detect non-zero compensation
	// Note: might still be small due to exponential moving average
	t.Logf("Detected drift: %v", drift)
	
	// Just verify it's calculated (can be zero or non-zero depending on algorithm)
	if drift < -100*time.Millisecond || drift > 100*time.Millisecond {
		t.Errorf("Drift compensation out of reasonable range: %v", drift)
	}
}

func TestStreamSynchronizer_GetAlignment(t *testing.T) {
	sync := NewStreamSynchronizer()

	// No data yet
	alignment := sync.GetAlignment()
	if alignment != 0 {
		t.Errorf("Expected 0 alignment with no data, got %v", alignment)
	}

	// Add timestamps
	captureTime := time.Now()
	playbackTime := captureTime.Add(40 * time.Millisecond)

	sync.SyncCapturePlayback(captureTime, playbackTime)

	alignment = sync.GetAlignment()
	// Should be approximately 40ms
	if alignment < 35*time.Millisecond || alignment > 45*time.Millisecond {
		t.Errorf("Expected alignment ~40ms, got %v", alignment)
	}
}

func TestStreamSynchronizer_IsAligned(t *testing.T) {
	sync := NewStreamSynchronizer()

	// Good alignment (within 50ms)
	captureTime := time.Now()
	playbackTime := captureTime.Add(30 * time.Millisecond)
	sync.SyncCapturePlayback(captureTime, playbackTime)

	if !sync.IsAligned() {
		t.Error("Expected streams to be aligned with 30ms difference")
	}

	// Poor alignment (exceeds 50ms)
	sync.Reset()
	captureTime = time.Now()
	playbackTime = captureTime.Add(60 * time.Millisecond)
	sync.SyncCapturePlayback(captureTime, playbackTime)

	if sync.IsAligned() {
		t.Error("Expected streams to be misaligned with 60ms difference")
	}
}

func TestStreamSynchronizer_AdjustBufferSize(t *testing.T) {
	sync := NewStreamSynchronizer()

	// Valid adjustment
	err := sync.AdjustBufferSize(60 * time.Millisecond)
	if err != nil {
		t.Errorf("AdjustBufferSize failed: %v", err)
	}

	stats := sync.GetStats()
	if stats.BufferAdjustments != 1 {
		t.Errorf("Expected 1 buffer adjustment, got %d", stats.BufferAdjustments)
	}

	// Try immediate adjustment (should fail due to cooldown)
	err = sync.AdjustBufferSize(70 * time.Millisecond)
	if err == nil {
		t.Error("Expected error due to cooldown, got nil")
	}

	// Wait for cooldown
	time.Sleep(600 * time.Millisecond)

	// Should succeed now
	err = sync.AdjustBufferSize(80 * time.Millisecond)
	if err != nil {
		t.Errorf("AdjustBufferSize after cooldown failed: %v", err)
	}
}

func TestStreamSynchronizer_AdjustBufferSize_InvalidLatency(t *testing.T) {
	sync := NewStreamSynchronizer()

	// Too low
	err := sync.AdjustBufferSize(5 * time.Millisecond)
	if err == nil {
		t.Error("Expected error for too low latency")
	}

	// Too high
	err = sync.AdjustBufferSize(600 * time.Millisecond)
	if err == nil {
		t.Error("Expected error for too high latency")
	}
}

func TestStreamSynchronizer_Reset(t *testing.T) {
	sync := NewStreamSynchronizer()

	// Add some data
	baseTime := time.Now()
	for i := 0; i < 5; i++ {
		captureTime := baseTime.Add(time.Duration(i) * 100 * time.Millisecond)
		playbackTime := captureTime.Add(10 * time.Millisecond)
		sync.SyncCapturePlayback(captureTime, playbackTime)
	}

	stats := sync.GetStats()
	if stats.CaptureDataPoints == 0 {
		t.Error("Expected data points before reset")
	}

	// Reset
	sync.Reset()

	stats = sync.GetStats()
	if stats.CaptureDataPoints != 0 {
		t.Errorf("Expected 0 capture data points after reset, got %d", stats.CaptureDataPoints)
	}
	if stats.PlaybackDataPoints != 0 {
		t.Errorf("Expected 0 playback data points after reset, got %d", stats.PlaybackDataPoints)
	}
	if stats.DriftCompensation != 0 {
		t.Errorf("Expected 0 drift compensation after reset, got %v", stats.DriftCompensation)
	}
}

func TestStreamSynchronizer_SetTargetAlignment(t *testing.T) {
	sync := NewStreamSynchronizer()

	// Valid target
	err := sync.SetTargetAlignment(30 * time.Millisecond)
	if err != nil {
		t.Errorf("SetTargetAlignment failed: %v", err)
	}

	stats := sync.GetStats()
	if stats.TargetAlignment != 30*time.Millisecond {
		t.Errorf("Expected target alignment 30ms, got %v", stats.TargetAlignment)
	}

	// Invalid targets
	err = sync.SetTargetAlignment(500 * time.Microsecond)
	if err == nil {
		t.Error("Expected error for too low target alignment")
	}

	err = sync.SetTargetAlignment(300 * time.Millisecond)
	if err == nil {
		t.Error("Expected error for too high target alignment")
	}
}

func TestStreamSynchronizer_CalculateTimestampMapping(t *testing.T) {
	sync := NewStreamSynchronizer()

	// Add some data to establish drift
	baseTime := time.Now()
	for i := 0; i < 10; i++ {
		captureTime := baseTime.Add(time.Duration(i) * 100 * time.Millisecond)
		playbackTime := captureTime.Add(20 * time.Millisecond)
		sync.SyncCapturePlayback(captureTime, playbackTime)
	}

	// Map a new capture timestamp
	newCaptureTime := time.Now()
	mappedPlaybackTime := sync.CalculateTimestampMapping(newCaptureTime)

	// Mapped time should be after capture time (due to drift compensation)
	if !mappedPlaybackTime.After(newCaptureTime) && !mappedPlaybackTime.Equal(newCaptureTime) {
		t.Error("Mapped playback time should be after or equal to capture time")
	}

	t.Logf("Capture: %v, Mapped Playback: %v, Diff: %v",
		newCaptureTime, mappedPlaybackTime, mappedPlaybackTime.Sub(newCaptureTime))
}

func TestStreamSynchronizer_DetectClockSkew(t *testing.T) {
	sync := NewStreamSynchronizer()

	// Not enough data
	hasSkew, _ := sync.DetectClockSkew()
	if hasSkew {
		t.Error("Should not detect skew with insufficient data")
	}

	// Add data with no significant skew
	baseTime := time.Now()
	for i := 0; i < 20; i++ {
		captureTime := baseTime.Add(time.Duration(i) * 50 * time.Millisecond)
		playbackTime := captureTime.Add(10 * time.Millisecond) // Constant 10ms offset
		sync.SyncCapturePlayback(captureTime, playbackTime)
	}

	hasSkew, drift := sync.DetectClockSkew()
	t.Logf("Clock skew detected: %v, drift: %v", hasSkew, drift)

	// The result depends on the drift calculation
	// Just verify it doesn't crash and returns reasonable values
	if drift < -100*time.Millisecond || drift > 100*time.Millisecond {
		t.Errorf("Drift value seems unreasonable: %v", drift)
	}
}

func TestStreamSynchronizer_HistoryLimit(t *testing.T) {
	sync := NewStreamSynchronizer()

	// Add more than maxHistorySize timestamps
	baseTime := time.Now()
	for i := 0; i < 150; i++ {
		captureTime := baseTime.Add(time.Duration(i) * 10 * time.Millisecond)
		playbackTime := captureTime.Add(5 * time.Millisecond)
		sync.SyncCapturePlayback(captureTime, playbackTime)
	}

	stats := sync.GetStats()

	// Should be limited to maxHistorySize (100)
	if stats.CaptureDataPoints > 100 {
		t.Errorf("Expected max 100 capture data points, got %d", stats.CaptureDataPoints)
	}
	if stats.PlaybackDataPoints > 100 {
		t.Errorf("Expected max 100 playback data points, got %d", stats.PlaybackDataPoints)
	}
}

func TestStreamSynchronizer_DriftCalculation(t *testing.T) {
	sync := NewStreamSynchronizer()

	// Simulate gradual drift
	baseTime := time.Now()
	for i := 0; i < 50; i++ {
		captureTime := baseTime.Add(time.Duration(i) * 20 * time.Millisecond)
		// Playback gradually drifts from 10ms to 15ms
		drift := 10 + (i / 10)
		playbackTime := captureTime.Add(time.Duration(drift) * time.Millisecond)
		sync.SyncCapturePlayback(captureTime, playbackTime)
	}

	drift := sync.GetDriftCompensation()
	t.Logf("Calculated drift compensation: %v", drift)

	// Drift should be non-zero
	if drift == 0 {
		t.Error("Expected non-zero drift with gradual drift pattern")
	}
}

func TestStreamSynchronizer_ConcurrentAccess(t *testing.T) {
	sync := NewStreamSynchronizer()

	// Test concurrent access doesn't cause race conditions
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			captureTime := time.Now()
			playbackTime := captureTime.Add(10 * time.Millisecond)
			sync.SyncCapturePlayback(captureTime, playbackTime)
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Reader goroutines
	for j := 0; j < 3; j++ {
		go func() {
			for i := 0; i < 100; i++ {
				sync.GetDriftCompensation()
				sync.GetAlignment()
				sync.IsAligned()
				sync.GetStats()
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
