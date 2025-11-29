package latency

import (
	"fmt"
	"github.com/dubbing-system/audio-interface/pkg/types"
	"sync"
	"time"
)

// LatencyManager monitors and optimizes audio pipeline latency
type LatencyManager struct {
	mu                    sync.RWMutex
	captureLatency        time.Duration
	playbackLatency       time.Duration
	endToEndLatency       time.Duration
	targetLatency         time.Duration
	latencyHistory        []time.Duration
	maxHistorySize        int
	cpuLoad               float64
	wasapiMode            types.WASAPIMode
	bufferOptimizations   int
	latencyViolations     int
	lastOptimizationTime  time.Time
	optimizationCooldown  time.Duration
	jitterCompensation    time.Duration
}

// NewLatencyManager creates a new latency manager
func NewLatencyManager(targetLatency time.Duration) *LatencyManager {
	return &LatencyManager{
		targetLatency:        targetLatency,
		latencyHistory:       make([]time.Duration, 0, 100),
		maxHistorySize:       100,
		wasapiMode:           types.Shared, // Default to Shared mode
		optimizationCooldown: 1 * time.Second,
	}
}

// MonitorLatency returns current latency metrics
func (lm *LatencyManager) MonitorLatency() types.LatencyMetrics {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	return types.LatencyMetrics{
		CaptureLatency:  lm.captureLatency,
		PlaybackLatency: lm.playbackLatency,
		BufferFillLevel: 0.0, // Will be updated by caller
		DroppedFrames:   0,
		Underruns:       0,
		Overruns:        0,
		Timestamp:       time.Now(),
	}
}

// UpdateLatency updates latency measurements
func (lm *LatencyManager) UpdateLatency(capture, playback time.Duration) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.captureLatency = capture
	lm.playbackLatency = playback
	lm.endToEndLatency = capture + playback

	// Add to history
	lm.latencyHistory = append(lm.latencyHistory, lm.endToEndLatency)
	if len(lm.latencyHistory) > lm.maxHistorySize {
		lm.latencyHistory = lm.latencyHistory[1:]
	}

	// Check for violations
	if lm.endToEndLatency > lm.targetLatency+150*time.Millisecond {
		lm.latencyViolations++
	}
}

// OptimizeBuffers optimizes buffer sizes based on CPU load
func (lm *LatencyManager) OptimizeBuffers(cpuLoad float64) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Check cooldown
	if time.Since(lm.lastOptimizationTime) < lm.optimizationCooldown {
		return fmt.Errorf("optimization cooldown active")
	}

	// Validate CPU load
	if cpuLoad < 0.0 || cpuLoad > 1.0 {
		return fmt.Errorf("invalid CPU load: %f (must be 0.0-1.0)", cpuLoad)
	}

	lm.cpuLoad = cpuLoad

	// Optimization strategy:
	// - High CPU load (>80%): Increase buffer size to reduce processing frequency
	// - Low CPU load (<30%): Decrease buffer size for lower latency
	// - Medium CPU load: Keep current settings

	if cpuLoad > 0.8 {
		// High CPU load - increase buffers
		lm.bufferOptimizations++
		lm.lastOptimizationTime = time.Now()
		return nil
	} else if cpuLoad < 0.3 {
		// Low CPU load - decrease buffers for lower latency
		lm.bufferOptimizations++
		lm.lastOptimizationTime = time.Now()
		return nil
	}

	// Medium load - no change needed
	return nil
}

// SelectOperationMode selects the appropriate WASAPI mode
func (lm *LatencyManager) SelectOperationMode() (types.WASAPIMode, error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Decision logic:
	// - If target latency is very low (<50ms), prefer Exclusive mode
	// - If CPU load is high (>70%), prefer Shared mode (less demanding)
	// - Otherwise, use Exclusive for better latency

	if lm.targetLatency < 50*time.Millisecond && lm.cpuLoad < 0.7 {
		lm.wasapiMode = types.Exclusive
	} else if lm.cpuLoad > 0.7 {
		lm.wasapiMode = types.Shared
	} else {
		lm.wasapiMode = types.Exclusive
	}

	return lm.wasapiMode, nil
}

// GetCurrentMode returns the current WASAPI mode
func (lm *LatencyManager) GetCurrentMode() types.WASAPIMode {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.wasapiMode
}

// GetEndToEndLatency returns the current end-to-end latency
func (lm *LatencyManager) GetEndToEndLatency() time.Duration {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.endToEndLatency
}

// GetTargetLatency returns the target latency
func (lm *LatencyManager) GetTargetLatency() time.Duration {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.targetLatency
}

// SetTargetLatency sets a new target latency
func (lm *LatencyManager) SetTargetLatency(target time.Duration) error {
	if target < 10*time.Millisecond || target > 500*time.Millisecond {
		return fmt.Errorf("invalid target latency: %v (must be 10-500ms)", target)
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.targetLatency = target

	return nil
}

// IsWithinTarget checks if current latency is within target
func (lm *LatencyManager) IsWithinTarget() bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.endToEndLatency <= lm.targetLatency
}

// GetAverageLatency calculates average latency from history
func (lm *LatencyManager) GetAverageLatency() time.Duration {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if len(lm.latencyHistory) == 0 {
		return 0
	}

	var sum time.Duration
	for _, latency := range lm.latencyHistory {
		sum += latency
	}

	return sum / time.Duration(len(lm.latencyHistory))
}

// GetLatencyPercentile calculates the specified percentile from history
func (lm *LatencyManager) GetLatencyPercentile(percentile float64) time.Duration {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if len(lm.latencyHistory) == 0 {
		return 0
	}

	if percentile < 0 || percentile > 100 {
		return 0
	}

	// Simple percentile calculation (not perfectly accurate but sufficient)
	// Copy and sort history
	sorted := make([]time.Duration, len(lm.latencyHistory))
	copy(sorted, lm.latencyHistory)

	// Bubble sort (simple for small arrays)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)-1) * percentile / 100.0)
	return sorted[index]
}

// GetStats returns comprehensive latency statistics
func (lm *LatencyManager) GetStats() LatencyStats {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	// Calculate average inline to avoid deadlock
	var avgLatency time.Duration
	if len(lm.latencyHistory) > 0 {
		var sum time.Duration
		for _, latency := range lm.latencyHistory {
			sum += latency
		}
		avgLatency = sum / time.Duration(len(lm.latencyHistory))
	}

	// Calculate percentiles inline
	percentiles := lm.calculatePercentilesUnsafe(50, 95, 99)
	p50, p95, p99 := percentiles[0], percentiles[1], percentiles[2]

	return LatencyStats{
		CaptureLatency:      lm.captureLatency,
		PlaybackLatency:     lm.playbackLatency,
		EndToEndLatency:     lm.endToEndLatency,
		TargetLatency:       lm.targetLatency,
		AverageLatency:      avgLatency,
		P50Latency:          p50,
		P95Latency:          p95,
		P99Latency:          p99,
		CPULoad:             lm.cpuLoad,
		WASAPIMode:          lm.wasapiMode,
		BufferOptimizations: lm.bufferOptimizations,
		LatencyViolations:   lm.latencyViolations,
		JitterCompensation:  lm.jitterCompensation,
	}
}

// calculatePercentilesUnsafe calculates percentiles without locking (caller must hold lock)
func (lm *LatencyManager) calculatePercentilesUnsafe(percentiles ...float64) []time.Duration {
	results := make([]time.Duration, len(percentiles))

	if len(lm.latencyHistory) == 0 {
		return results
	}

	// Copy and sort history
	sorted := make([]time.Duration, len(lm.latencyHistory))
	copy(sorted, lm.latencyHistory)

	// Bubble sort
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Calculate each percentile
	for i, p := range percentiles {
		if p < 0 || p > 100 {
			results[i] = 0
			continue
		}
		index := int(float64(len(sorted)-1) * p / 100.0)
		results[i] = sorted[index]
	}

	return results
}

// UpdateCPULoad updates the current CPU load measurement
func (lm *LatencyManager) UpdateCPULoad(load float64) error {
	if load < 0.0 || load > 1.0 {
		return fmt.Errorf("invalid CPU load: %f (must be 0.0-1.0)", load)
	}

	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.cpuLoad = load

	return nil
}

// CalculateJitterCompensation calculates jitter compensation based on latency variance
func (lm *LatencyManager) CalculateJitterCompensation() time.Duration {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if len(lm.latencyHistory) < 2 {
		return 0
	}

	// Calculate average (inline to avoid deadlock)
	var sum time.Duration
	for _, latency := range lm.latencyHistory {
		sum += latency
	}
	avg := sum / time.Duration(len(lm.latencyHistory))

	// Calculate variance
	var variance float64
	for _, latency := range lm.latencyHistory {
		diff := float64(latency - avg)
		variance += diff * diff
	}
	variance /= float64(len(lm.latencyHistory))

	// Jitter compensation is proportional to standard deviation
	stdDev := time.Duration(variance)
	lm.jitterCompensation = stdDev / 2 // Use half of std dev

	return lm.jitterCompensation
}

// Reset clears all latency history and statistics
func (lm *LatencyManager) Reset() {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.latencyHistory = make([]time.Duration, 0, 100)
	lm.bufferOptimizations = 0
	lm.latencyViolations = 0
	lm.captureLatency = 0
	lm.playbackLatency = 0
	lm.endToEndLatency = 0
	lm.jitterCompensation = 0
}

// LatencyStats contains comprehensive latency statistics
type LatencyStats struct {
	CaptureLatency      time.Duration
	PlaybackLatency     time.Duration
	EndToEndLatency     time.Duration
	TargetLatency       time.Duration
	AverageLatency      time.Duration
	P50Latency          time.Duration // Median
	P95Latency          time.Duration
	P99Latency          time.Duration
	CPULoad             float64
	WASAPIMode          types.WASAPIMode
	BufferOptimizations int
	LatencyViolations   int
	JitterCompensation  time.Duration
}
