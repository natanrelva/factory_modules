package adaptive

import (
	"github.com/dubbing-system/audio-interface/pkg/types"
	"sync"
	"time"
)

// AdaptivePolicy implements adaptive optimization policies based on system metrics
type AdaptivePolicy struct {
	mu                   sync.RWMutex
	latencyThreshold     time.Duration
	bufferAdjustmentStep int
	cpuThreshold         float64
	underrunThreshold    int
	actionsApplied       int64
}

// Action represents an optimization action to be applied
type Action struct {
	Type  ActionType
	Value interface{}
}

// ActionType defines the type of optimization action
type ActionType int

const (
	// ReduceBuffer reduces buffer size to decrease latency
	ReduceBuffer ActionType = iota
	// IncreaseBuffer increases buffer size to prevent underruns
	IncreaseBuffer
	// SwitchToExclusiveMode switches to WASAPI exclusive mode for lower latency
	SwitchToExclusiveMode
	// SwitchToSharedMode switches to WASAPI shared mode for better compatibility
	SwitchToSharedMode
	// ApplyDriftCompensation applies clock drift compensation
	ApplyDriftCompensation
)

// String returns the string representation of an ActionType
func (at ActionType) String() string {
	switch at {
	case ReduceBuffer:
		return "ReduceBuffer"
	case IncreaseBuffer:
		return "IncreaseBuffer"
	case SwitchToExclusiveMode:
		return "SwitchToExclusiveMode"
	case SwitchToSharedMode:
		return "SwitchToSharedMode"
	case ApplyDriftCompensation:
		return "ApplyDriftCompensation"
	default:
		return "Unknown"
	}
}

// NewAdaptivePolicy creates a new adaptive policy with default thresholds
func NewAdaptivePolicy() *AdaptivePolicy {
	return &AdaptivePolicy{
		latencyThreshold:     80 * time.Millisecond,
		bufferAdjustmentStep: 2,
		cpuThreshold:         0.8,
		underrunThreshold:    5,
	}
}

// NewAdaptivePolicyWithThresholds creates a policy with custom thresholds
func NewAdaptivePolicyWithThresholds(latency time.Duration, bufferStep int, cpu float64, underruns int) *AdaptivePolicy {
	return &AdaptivePolicy{
		latencyThreshold:     latency,
		bufferAdjustmentStep: bufferStep,
		cpuThreshold:         cpu,
		underrunThreshold:    underruns,
	}
}

// Evaluate evaluates system metrics and returns recommended actions
func (ap *AdaptivePolicy) Evaluate(metrics types.LatencyMetrics, underruns int, cpuLoad float64) []Action {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	actions := make([]Action, 0)

	// Policy 1: High latency → Reduce buffer
	totalLatency := metrics.CaptureLatency + metrics.PlaybackLatency
	if totalLatency > ap.latencyThreshold {
		actions = append(actions, Action{
			Type:  ReduceBuffer,
			Value: ap.bufferAdjustmentStep,
		})
	}

	// Policy 2: Frequent underruns → Increase buffer
	if underruns > ap.underrunThreshold {
		actions = append(actions, Action{
			Type:  IncreaseBuffer,
			Value: ap.bufferAdjustmentStep,
		})
	}

	// Policy 3: High CPU → Switch to Shared mode
	if cpuLoad > ap.cpuThreshold {
		actions = append(actions, Action{
			Type: SwitchToSharedMode,
		})
	} else if cpuLoad < 0.5 {
		// Low CPU → Switch to Exclusive mode for better latency
		actions = append(actions, Action{
			Type: SwitchToExclusiveMode,
		})
	}

	// Policy 4: Significant drift → Apply compensation
	if metrics.DroppedFrames > 0 {
		actions = append(actions, Action{
			Type: ApplyDriftCompensation,
		})
	}

	ap.actionsApplied += int64(len(actions))

	return actions
}

// EvaluateWithMetrics evaluates using only LatencyMetrics (simplified version)
func (ap *AdaptivePolicy) EvaluateWithMetrics(metrics types.LatencyMetrics) []Action {
	return ap.Evaluate(metrics, metrics.Underruns, 0.5) // Default CPU load
}

// SetLatencyThreshold sets the latency threshold for triggering buffer reduction
func (ap *AdaptivePolicy) SetLatencyThreshold(threshold time.Duration) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	ap.latencyThreshold = threshold
}

// GetLatencyThreshold returns the current latency threshold
func (ap *AdaptivePolicy) GetLatencyThreshold() time.Duration {
	ap.mu.RLock()
	defer ap.mu.RUnlock()
	return ap.latencyThreshold
}

// SetBufferAdjustmentStep sets the buffer adjustment step size
func (ap *AdaptivePolicy) SetBufferAdjustmentStep(step int) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	ap.bufferAdjustmentStep = step
}

// GetBufferAdjustmentStep returns the current buffer adjustment step
func (ap *AdaptivePolicy) GetBufferAdjustmentStep() int {
	ap.mu.RLock()
	defer ap.mu.RUnlock()
	return ap.bufferAdjustmentStep
}

// SetCPUThreshold sets the CPU threshold for mode switching
func (ap *AdaptivePolicy) SetCPUThreshold(threshold float64) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	ap.cpuThreshold = threshold
}

// GetCPUThreshold returns the current CPU threshold
func (ap *AdaptivePolicy) GetCPUThreshold() float64 {
	ap.mu.RLock()
	defer ap.mu.RUnlock()
	return ap.cpuThreshold
}

// SetUnderrunThreshold sets the underrun threshold for buffer increase
func (ap *AdaptivePolicy) SetUnderrunThreshold(threshold int) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	ap.underrunThreshold = threshold
}

// GetUnderrunThreshold returns the current underrun threshold
func (ap *AdaptivePolicy) GetUnderrunThreshold() int {
	ap.mu.RLock()
	defer ap.mu.RUnlock()
	return ap.underrunThreshold
}

// GetActionsApplied returns the total number of actions applied
func (ap *AdaptivePolicy) GetActionsApplied() int64 {
	ap.mu.RLock()
	defer ap.mu.RUnlock()
	return ap.actionsApplied
}

// Reset resets statistics
func (ap *AdaptivePolicy) Reset() {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	ap.actionsApplied = 0
}

// GetStats returns policy statistics
func (ap *AdaptivePolicy) GetStats() PolicyStats {
	ap.mu.RLock()
	defer ap.mu.RUnlock()

	return PolicyStats{
		LatencyThreshold:     ap.latencyThreshold,
		BufferAdjustmentStep: ap.bufferAdjustmentStep,
		CPUThreshold:         ap.cpuThreshold,
		UnderrunThreshold:    ap.underrunThreshold,
		ActionsApplied:       ap.actionsApplied,
	}
}

// PolicyStats contains policy statistics
type PolicyStats struct {
	LatencyThreshold     time.Duration
	BufferAdjustmentStep int
	CPUThreshold         float64
	UnderrunThreshold    int
	ActionsApplied       int64
}
