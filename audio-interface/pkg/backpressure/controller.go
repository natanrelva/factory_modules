package backpressure

import (
	"sync"
	"time"
)

// BackpressureController manages flow control using hysteresis-based watermarks
type BackpressureController struct {
	mu                 sync.RWMutex
	highWaterMark      float64
	lowWaterMark       float64
	currentFillLevel   float64
	backpressureActive bool
	eventsCount        int64
	throttlingDuration time.Duration
}

// NewBackpressureController creates a new backpressure controller with default watermarks
func NewBackpressureController() *BackpressureController {
	return &BackpressureController{
		highWaterMark: 0.8, // 80%
		lowWaterMark:  0.2, // 20%
	}
}

// NewBackpressureControllerWithWatermarks creates a controller with custom watermarks
func NewBackpressureControllerWithWatermarks(high, low float64) *BackpressureController {
	return &BackpressureController{
		highWaterMark: high,
		lowWaterMark:  low,
	}
}

// ShouldApplyBackpressure determines if backpressure should be applied based on fill level
// Uses hysteresis to avoid oscillation
func (bp *BackpressureController) ShouldApplyBackpressure(fillLevel float64) bool {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.currentFillLevel = fillLevel

	// Hysteresis: activate at high watermark, deactivate at low watermark
	if fillLevel > bp.highWaterMark {
		if !bp.backpressureActive {
			bp.eventsCount++
		}
		bp.backpressureActive = true
	} else if fillLevel < bp.lowWaterMark {
		bp.backpressureActive = false
	}

	return bp.backpressureActive
}

// GetThrottleDuration returns the recommended throttle duration based on fill level
func (bp *BackpressureController) GetThrottleDuration() time.Duration {
	bp.mu.RLock()
	defer bp.mu.RUnlock()

	// Throttle duration based on fill level
	if bp.currentFillLevel > 0.9 {
		return 10 * time.Millisecond
	} else if bp.currentFillLevel > 0.8 {
		return 5 * time.Millisecond
	}
	return 0
}

// GetStats returns backpressure statistics
func (bp *BackpressureController) GetStats() (events int64, duration time.Duration) {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	return bp.eventsCount, bp.throttlingDuration
}

// IsActive returns whether backpressure is currently active
func (bp *BackpressureController) IsActive() bool {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	return bp.backpressureActive
}

// GetCurrentFillLevel returns the current fill level
func (bp *BackpressureController) GetCurrentFillLevel() float64 {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	return bp.currentFillLevel
}

// Reset resets all statistics
func (bp *BackpressureController) Reset() {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.eventsCount = 0
	bp.throttlingDuration = 0
	bp.backpressureActive = false
	bp.currentFillLevel = 0.0
}

// SetWatermarks updates the high and low watermarks
func (bp *BackpressureController) SetWatermarks(high, low float64) error {
	if high <= low {
		return ErrInvalidWatermarks
	}
	if high > 1.0 || low < 0.0 {
		return ErrInvalidWatermarks
	}

	bp.mu.Lock()
	defer bp.mu.Unlock()

	bp.highWaterMark = high
	bp.lowWaterMark = low

	return nil
}

// GetWatermarks returns the current watermark settings
func (bp *BackpressureController) GetWatermarks() (high, low float64) {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	return bp.highWaterMark, bp.lowWaterMark
}

// RecordThrottling records a throttling event
func (bp *BackpressureController) RecordThrottling(duration time.Duration) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.throttlingDuration += duration
}

// ErrInvalidWatermarks is returned when watermark values are invalid
var ErrInvalidWatermarks = &BackpressureError{msg: "invalid watermarks: high must be > low and both in range [0,1]"}

// BackpressureError represents a backpressure-related error
type BackpressureError struct {
	msg string
}

func (e *BackpressureError) Error() string {
	return e.msg
}
