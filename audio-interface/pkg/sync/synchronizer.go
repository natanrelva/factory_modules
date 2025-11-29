package sync

import (
	"fmt"
	"sync"
	"time"
)

// StreamSynchronizer manages temporal alignment between capture and playback streams
type StreamSynchronizer struct {
	mu                  sync.RWMutex
	captureTimestamps   []time.Time
	playbackTimestamps  []time.Time
	driftCompensation   time.Duration
	maxHistorySize      int
	targetAlignment     time.Duration // Target alignment accuracy (50ms)
	bufferAdjustments   int
	lastAdjustmentTime  time.Time
	adjustmentCooldown  time.Duration
}

// NewStreamSynchronizer creates a new stream synchronizer
func NewStreamSynchronizer() *StreamSynchronizer {
	return &StreamSynchronizer{
		captureTimestamps:   make([]time.Time, 0, 100),
		playbackTimestamps:  make([]time.Time, 0, 100),
		maxHistorySize:      100,
		targetAlignment:     50 * time.Millisecond,
		adjustmentCooldown:  500 * time.Millisecond, // Don't adjust too frequently
	}
}

// SyncCapturePlayback synchronizes capture and playback timestamps
func (s *StreamSynchronizer) SyncCapturePlayback(captureTime, playbackTime time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Add timestamps to history
	s.captureTimestamps = append(s.captureTimestamps, captureTime)
	s.playbackTimestamps = append(s.playbackTimestamps, playbackTime)

	// Trim history if too large
	if len(s.captureTimestamps) > s.maxHistorySize {
		s.captureTimestamps = s.captureTimestamps[1:]
	}
	if len(s.playbackTimestamps) > s.maxHistorySize {
		s.playbackTimestamps = s.playbackTimestamps[1:]
	}

	// Calculate drift
	s.calculateDrift()

	return nil
}

// calculateDrift calculates clock drift between capture and playback
func (s *StreamSynchronizer) calculateDrift() {
	if len(s.captureTimestamps) < 2 || len(s.playbackTimestamps) < 2 {
		return
	}

	// Calculate time difference between first and last timestamps
	captureSpan := s.captureTimestamps[len(s.captureTimestamps)-1].Sub(s.captureTimestamps[0])
	playbackSpan := s.playbackTimestamps[len(s.playbackTimestamps)-1].Sub(s.playbackTimestamps[0])

	// Drift is the difference between the two spans
	drift := playbackSpan - captureSpan

	// Update drift compensation (use exponential moving average)
	alpha := 0.1 // Smoothing factor
	s.driftCompensation = time.Duration(float64(s.driftCompensation)*(1-alpha) + float64(drift)*alpha)
}

// GetDriftCompensation returns the current drift compensation value
func (s *StreamSynchronizer) GetDriftCompensation() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.driftCompensation
}

// AdjustBufferSize adjusts buffer size based on target latency
func (s *StreamSynchronizer) AdjustBufferSize(targetLatency time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check cooldown period
	if time.Since(s.lastAdjustmentTime) < s.adjustmentCooldown {
		return fmt.Errorf("adjustment cooldown active")
	}

	// Validate target latency
	if targetLatency < 10*time.Millisecond {
		return fmt.Errorf("target latency too low: %v", targetLatency)
	}
	if targetLatency > 500*time.Millisecond {
		return fmt.Errorf("target latency too high: %v", targetLatency)
	}

	s.bufferAdjustments++
	s.lastAdjustmentTime = time.Now()

	return nil
}

// GetAlignment returns the current alignment between capture and playback
func (s *StreamSynchronizer) GetAlignment() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.captureTimestamps) == 0 || len(s.playbackTimestamps) == 0 {
		return 0
	}

	// Calculate alignment as the difference between latest timestamps
	lastCapture := s.captureTimestamps[len(s.captureTimestamps)-1]
	lastPlayback := s.playbackTimestamps[len(s.playbackTimestamps)-1]

	alignment := lastPlayback.Sub(lastCapture)
	if alignment < 0 {
		alignment = -alignment
	}

	return alignment
}

// IsAligned checks if streams are aligned within target accuracy
func (s *StreamSynchronizer) IsAligned() bool {
	alignment := s.GetAlignment()
	return alignment <= s.targetAlignment
}

// GetStats returns synchronization statistics
func (s *StreamSynchronizer) GetStats() SyncStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return SyncStats{
		DriftCompensation:  s.driftCompensation,
		CurrentAlignment:   s.GetAlignment(),
		TargetAlignment:    s.targetAlignment,
		BufferAdjustments:  s.bufferAdjustments,
		CaptureDataPoints:  len(s.captureTimestamps),
		PlaybackDataPoints: len(s.playbackTimestamps),
	}
}

// Reset clears all synchronization history
func (s *StreamSynchronizer) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.captureTimestamps = make([]time.Time, 0, 100)
	s.playbackTimestamps = make([]time.Time, 0, 100)
	s.driftCompensation = 0
	s.bufferAdjustments = 0
}

// SetTargetAlignment sets the target alignment accuracy
func (s *StreamSynchronizer) SetTargetAlignment(target time.Duration) error {
	if target < time.Millisecond || target > 200*time.Millisecond {
		return fmt.Errorf("invalid target alignment: %v (must be 1-200ms)", target)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.targetAlignment = target

	return nil
}

// CalculateTimestampMapping maps a capture timestamp to expected playback timestamp
func (s *StreamSynchronizer) CalculateTimestampMapping(captureTime time.Time) time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Apply drift compensation
	playbackTime := captureTime.Add(s.driftCompensation)

	return playbackTime
}

// DetectClockSkew detects if there's significant clock skew between streams
func (s *StreamSynchronizer) DetectClockSkew() (bool, time.Duration) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Threshold for significant skew: 2ms per second
	threshold := 2 * time.Millisecond

	if len(s.captureTimestamps) < 10 {
		return false, 0
	}

	// Calculate drift rate
	captureSpan := s.captureTimestamps[len(s.captureTimestamps)-1].Sub(s.captureTimestamps[0])
	if captureSpan == 0 {
		return false, 0
	}

	driftRate := s.driftCompensation / captureSpan

	hasSkew := driftRate > threshold || driftRate < -threshold

	return hasSkew, s.driftCompensation
}

// SyncStats contains synchronization statistics
type SyncStats struct {
	DriftCompensation  time.Duration
	CurrentAlignment   time.Duration
	TargetAlignment    time.Duration
	BufferAdjustments  int
	CaptureDataPoints  int
	PlaybackDataPoints int
}
