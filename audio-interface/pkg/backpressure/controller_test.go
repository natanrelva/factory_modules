package backpressure

import (
	"testing"
	"time"
)

func TestNewBackpressureController(t *testing.T) {
	bp := NewBackpressureController()
	if bp == nil {
		t.Fatal("NewBackpressureController returned nil")
	}

	high, low := bp.GetWatermarks()
	if high != 0.8 {
		t.Errorf("Expected high watermark 0.8, got %f", high)
	}
	if low != 0.2 {
		t.Errorf("Expected low watermark 0.2, got %f", low)
	}
}

func TestNewBackpressureControllerWithWatermarks(t *testing.T) {
	bp := NewBackpressureControllerWithWatermarks(0.9, 0.1)
	
	high, low := bp.GetWatermarks()
	if high != 0.9 {
		t.Errorf("Expected high watermark 0.9, got %f", high)
	}
	if low != 0.1 {
		t.Errorf("Expected low watermark 0.1, got %f", low)
	}
}

func TestBackpressureController_ShouldApplyBackpressure(t *testing.T) {
	bp := NewBackpressureController()

	// Initially no backpressure
	if bp.ShouldApplyBackpressure(0.5) {
		t.Error("Should not apply backpressure at 50% fill")
	}

	// Exceed high watermark - should activate
	if !bp.ShouldApplyBackpressure(0.85) {
		t.Error("Should apply backpressure at 85% fill")
	}

	// Stay above high watermark - should remain active
	if !bp.ShouldApplyBackpressure(0.82) {
		t.Error("Should keep backpressure active at 82% fill")
	}

	// Drop below low watermark - should deactivate
	if bp.ShouldApplyBackpressure(0.15) {
		t.Error("Should not apply backpressure at 15% fill")
	}

	// Stay below low watermark - should remain inactive
	if bp.ShouldApplyBackpressure(0.18) {
		t.Error("Should keep backpressure inactive at 18% fill")
	}
}

func TestBackpressureController_Hysteresis(t *testing.T) {
	bp := NewBackpressureController()

	// Test hysteresis behavior
	tests := []struct {
		fillLevel float64
		expected  bool
		desc      string
	}{
		{0.5, false, "Below high watermark - inactive"},
		{0.85, true, "Above high watermark - activate"},
		{0.75, true, "Between watermarks - stay active"},
		{0.5, true, "Between watermarks - stay active"},
		{0.15, false, "Below low watermark - deactivate"},
		{0.5, false, "Between watermarks - stay inactive"},
		{0.75, false, "Between watermarks - stay inactive"},
		{0.85, true, "Above high watermark - activate again"},
	}

	for _, tt := range tests {
		result := bp.ShouldApplyBackpressure(tt.fillLevel)
		if result != tt.expected {
			t.Errorf("%s: fillLevel=%.2f, expected=%v, got=%v", 
				tt.desc, tt.fillLevel, tt.expected, result)
		}
	}
}

func TestBackpressureController_GetThrottleDuration(t *testing.T) {
	bp := NewBackpressureController()

	tests := []struct {
		fillLevel float64
		expected  time.Duration
	}{
		{0.5, 0},
		{0.75, 0},
		{0.85, 5 * time.Millisecond},
		{0.95, 10 * time.Millisecond},
	}

	for _, tt := range tests {
		bp.ShouldApplyBackpressure(tt.fillLevel)
		duration := bp.GetThrottleDuration()
		if duration != tt.expected {
			t.Errorf("fillLevel=%.2f: expected duration %v, got %v", 
				tt.fillLevel, tt.expected, duration)
		}
	}
}

func TestBackpressureController_GetStats(t *testing.T) {
	bp := NewBackpressureController()

	// Initially no events
	events, duration := bp.GetStats()
	if events != 0 {
		t.Errorf("Expected 0 events, got %d", events)
	}
	if duration != 0 {
		t.Errorf("Expected 0 duration, got %v", duration)
	}

	// Trigger backpressure events
	bp.ShouldApplyBackpressure(0.85) // First activation
	bp.ShouldApplyBackpressure(0.15) // Deactivation
	bp.ShouldApplyBackpressure(0.90) // Second activation

	events, _ = bp.GetStats()
	if events != 2 {
		t.Errorf("Expected 2 events, got %d", events)
	}
}

func TestBackpressureController_IsActive(t *testing.T) {
	bp := NewBackpressureController()

	if bp.IsActive() {
		t.Error("Should not be active initially")
	}

	bp.ShouldApplyBackpressure(0.85)
	if !bp.IsActive() {
		t.Error("Should be active after exceeding high watermark")
	}

	bp.ShouldApplyBackpressure(0.15)
	if bp.IsActive() {
		t.Error("Should not be active after dropping below low watermark")
	}
}

func TestBackpressureController_GetCurrentFillLevel(t *testing.T) {
	bp := NewBackpressureController()

	bp.ShouldApplyBackpressure(0.75)
	fillLevel := bp.GetCurrentFillLevel()
	if fillLevel != 0.75 {
		t.Errorf("Expected fill level 0.75, got %f", fillLevel)
	}
}

func TestBackpressureController_Reset(t *testing.T) {
	bp := NewBackpressureController()

	// Generate some activity
	bp.ShouldApplyBackpressure(0.85)
	bp.RecordThrottling(10 * time.Millisecond)

	events, duration := bp.GetStats()
	if events == 0 || duration == 0 {
		t.Error("Expected non-zero stats before reset")
	}

	// Reset
	bp.Reset()

	events, duration = bp.GetStats()
	if events != 0 {
		t.Errorf("Expected 0 events after reset, got %d", events)
	}
	if duration != 0 {
		t.Errorf("Expected 0 duration after reset, got %v", duration)
	}
	if bp.IsActive() {
		t.Error("Should not be active after reset")
	}
}

func TestBackpressureController_SetWatermarks(t *testing.T) {
	bp := NewBackpressureController()

	// Valid watermarks
	err := bp.SetWatermarks(0.9, 0.1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	high, low := bp.GetWatermarks()
	if high != 0.9 || low != 0.1 {
		t.Errorf("Watermarks not set correctly: high=%f, low=%f", high, low)
	}

	// Invalid watermarks - high <= low
	err = bp.SetWatermarks(0.5, 0.5)
	if err == nil {
		t.Error("Expected error for high <= low")
	}

	err = bp.SetWatermarks(0.3, 0.7)
	if err == nil {
		t.Error("Expected error for high < low")
	}

	// Invalid watermarks - out of range
	err = bp.SetWatermarks(1.5, 0.5)
	if err == nil {
		t.Error("Expected error for high > 1.0")
	}

	err = bp.SetWatermarks(0.5, -0.1)
	if err == nil {
		t.Error("Expected error for low < 0.0")
	}
}

func TestBackpressureController_RecordThrottling(t *testing.T) {
	bp := NewBackpressureController()

	bp.RecordThrottling(5 * time.Millisecond)
	bp.RecordThrottling(10 * time.Millisecond)

	_, duration := bp.GetStats()
	if duration != 15*time.Millisecond {
		t.Errorf("Expected total duration 15ms, got %v", duration)
	}
}

func TestBackpressureController_ConcurrentAccess(t *testing.T) {
	bp := NewBackpressureController()

	done := make(chan bool)

	// Multiple goroutines accessing controller
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				fillLevel := float64(j) / 100.0
				bp.ShouldApplyBackpressure(fillLevel)
				bp.GetThrottleDuration()
				bp.IsActive()
				bp.GetCurrentFillLevel()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic or deadlock
	events, _ := bp.GetStats()
	t.Logf("Concurrent test completed with %d events", events)
}

func TestBackpressureController_EdgeCases(t *testing.T) {
	bp := NewBackpressureController()

	// Test at exact watermark boundaries
	// Note: activation happens when fillLevel > highWaterMark (not >=)
	bp.ShouldApplyBackpressure(0.81) // Just above high watermark
	if !bp.IsActive() {
		t.Error("Should activate just above high watermark")
	}

	bp.Reset()
	bp.ShouldApplyBackpressure(0.19) // Just below low watermark
	if bp.IsActive() {
		t.Error("Should not activate just below low watermark")
	}

	// Test extreme values
	bp.ShouldApplyBackpressure(0.0)
	bp.ShouldApplyBackpressure(1.0)
	
	// Should not panic
}
