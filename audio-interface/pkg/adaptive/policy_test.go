package adaptive

import (
	"github.com/dubbing-system/audio-interface/pkg/types"
	"testing"
	"time"
)

func TestNewAdaptivePolicy(t *testing.T) {
	policy := NewAdaptivePolicy()
	if policy == nil {
		t.Fatal("NewAdaptivePolicy returned nil")
	}

	if policy.GetLatencyThreshold() != 80*time.Millisecond {
		t.Errorf("Expected latency threshold 80ms, got %v", policy.GetLatencyThreshold())
	}
}

func TestNewAdaptivePolicyWithThresholds(t *testing.T) {
	policy := NewAdaptivePolicyWithThresholds(
		100*time.Millisecond, // latency
		3,                    // buffer step
		0.9,                  // CPU
		10,                   // underruns
	)

	if policy.GetLatencyThreshold() != 100*time.Millisecond {
		t.Errorf("Expected latency threshold 100ms, got %v", policy.GetLatencyThreshold())
	}
	if policy.GetBufferAdjustmentStep() != 3 {
		t.Errorf("Expected buffer step 3, got %d", policy.GetBufferAdjustmentStep())
	}
	if policy.GetCPUThreshold() != 0.9 {
		t.Errorf("Expected CPU threshold 0.9, got %f", policy.GetCPUThreshold())
	}
	if policy.GetUnderrunThreshold() != 10 {
		t.Errorf("Expected underrun threshold 10, got %d", policy.GetUnderrunThreshold())
	}
}

func TestAdaptivePolicy_EvaluateHighLatency(t *testing.T) {
	policy := NewAdaptivePolicy()

	metrics := types.LatencyMetrics{
		CaptureLatency:  50 * time.Millisecond,
		PlaybackLatency: 40 * time.Millisecond, // Total: 90ms > 80ms threshold
		Underruns:       0,
	}

	actions := policy.Evaluate(metrics, 0, 0.5)

	if len(actions) == 0 {
		t.Fatal("Expected at least one action for high latency")
	}

	foundReduceBuffer := false
	for _, action := range actions {
		if action.Type == ReduceBuffer {
			foundReduceBuffer = true
			if action.Value.(int) != 2 {
				t.Errorf("Expected buffer reduction of 2, got %v", action.Value)
			}
		}
	}

	if !foundReduceBuffer {
		t.Error("Expected ReduceBuffer action for high latency")
	}
}

func TestAdaptivePolicy_EvaluateFrequentUnderruns(t *testing.T) {
	policy := NewAdaptivePolicy()

	metrics := types.LatencyMetrics{
		CaptureLatency:  20 * time.Millisecond,
		PlaybackLatency: 30 * time.Millisecond,
		Underruns:       10, // > 5 threshold
	}

	actions := policy.Evaluate(metrics, 10, 0.5)

	foundIncreaseBuffer := false
	for _, action := range actions {
		if action.Type == IncreaseBuffer {
			foundIncreaseBuffer = true
			if action.Value.(int) != 2 {
				t.Errorf("Expected buffer increase of 2, got %v", action.Value)
			}
		}
	}

	if !foundIncreaseBuffer {
		t.Error("Expected IncreaseBuffer action for frequent underruns")
	}
}

func TestAdaptivePolicy_EvaluateHighCPU(t *testing.T) {
	policy := NewAdaptivePolicy()

	metrics := types.LatencyMetrics{
		CaptureLatency:  20 * time.Millisecond,
		PlaybackLatency: 30 * time.Millisecond,
	}

	actions := policy.Evaluate(metrics, 0, 0.9) // High CPU load

	foundSwitchToShared := false
	for _, action := range actions {
		if action.Type == SwitchToSharedMode {
			foundSwitchToShared = true
		}
	}

	if !foundSwitchToShared {
		t.Error("Expected SwitchToSharedMode action for high CPU")
	}
}

func TestAdaptivePolicy_EvaluateLowCPU(t *testing.T) {
	policy := NewAdaptivePolicy()

	metrics := types.LatencyMetrics{
		CaptureLatency:  20 * time.Millisecond,
		PlaybackLatency: 30 * time.Millisecond,
	}

	actions := policy.Evaluate(metrics, 0, 0.3) // Low CPU load

	foundSwitchToExclusive := false
	for _, action := range actions {
		if action.Type == SwitchToExclusiveMode {
			foundSwitchToExclusive = true
		}
	}

	if !foundSwitchToExclusive {
		t.Error("Expected SwitchToExclusiveMode action for low CPU")
	}
}

func TestAdaptivePolicy_EvaluateDroppedFrames(t *testing.T) {
	policy := NewAdaptivePolicy()

	metrics := types.LatencyMetrics{
		CaptureLatency:  20 * time.Millisecond,
		PlaybackLatency: 30 * time.Millisecond,
		DroppedFrames:   5,
	}

	actions := policy.Evaluate(metrics, 0, 0.5)

	foundDriftCompensation := false
	for _, action := range actions {
		if action.Type == ApplyDriftCompensation {
			foundDriftCompensation = true
		}
	}

	if !foundDriftCompensation {
		t.Error("Expected ApplyDriftCompensation action for dropped frames")
	}
}

func TestAdaptivePolicy_EvaluateMultipleConditions(t *testing.T) {
	policy := NewAdaptivePolicy()

	metrics := types.LatencyMetrics{
		CaptureLatency:  50 * time.Millisecond,
		PlaybackLatency: 40 * time.Millisecond, // High latency
		Underruns:       10,                     // Frequent underruns
		DroppedFrames:   5,                      // Dropped frames
	}

	actions := policy.Evaluate(metrics, 10, 0.9) // High CPU

	// Should have multiple actions
	if len(actions) < 2 {
		t.Errorf("Expected multiple actions, got %d", len(actions))
	}

	// Check for expected action types
	actionTypes := make(map[ActionType]bool)
	for _, action := range actions {
		actionTypes[action.Type] = true
	}

	// Note: ReduceBuffer and IncreaseBuffer are conflicting, so only one should be present
	// In this case, underruns take priority
	if !actionTypes[IncreaseBuffer] {
		t.Error("Expected IncreaseBuffer action")
	}
	if !actionTypes[SwitchToSharedMode] {
		t.Error("Expected SwitchToSharedMode action")
	}
	if !actionTypes[ApplyDriftCompensation] {
		t.Error("Expected ApplyDriftCompensation action")
	}
}

func TestAdaptivePolicy_EvaluateWithMetrics(t *testing.T) {
	policy := NewAdaptivePolicy()

	metrics := types.LatencyMetrics{
		CaptureLatency:  50 * time.Millisecond,
		PlaybackLatency: 40 * time.Millisecond,
		Underruns:       10,
	}

	actions := policy.EvaluateWithMetrics(metrics)

	if len(actions) == 0 {
		t.Error("Expected at least one action")
	}
}

func TestAdaptivePolicy_SettersGetters(t *testing.T) {
	policy := NewAdaptivePolicy()

	// Latency threshold
	policy.SetLatencyThreshold(100 * time.Millisecond)
	if policy.GetLatencyThreshold() != 100*time.Millisecond {
		t.Error("Latency threshold not set correctly")
	}

	// Buffer adjustment step
	policy.SetBufferAdjustmentStep(5)
	if policy.GetBufferAdjustmentStep() != 5 {
		t.Error("Buffer adjustment step not set correctly")
	}

	// CPU threshold
	policy.SetCPUThreshold(0.75)
	if policy.GetCPUThreshold() != 0.75 {
		t.Error("CPU threshold not set correctly")
	}

	// Underrun threshold
	policy.SetUnderrunThreshold(8)
	if policy.GetUnderrunThreshold() != 8 {
		t.Error("Underrun threshold not set correctly")
	}
}

func TestAdaptivePolicy_GetActionsApplied(t *testing.T) {
	policy := NewAdaptivePolicy()

	if policy.GetActionsApplied() != 0 {
		t.Error("Expected 0 actions applied initially")
	}

	metrics := types.LatencyMetrics{
		CaptureLatency:  50 * time.Millisecond,
		PlaybackLatency: 40 * time.Millisecond,
	}

	actions := policy.Evaluate(metrics, 0, 0.5)
	actionsCount := int64(len(actions))

	if policy.GetActionsApplied() != actionsCount {
		t.Errorf("Expected %d actions applied, got %d", actionsCount, policy.GetActionsApplied())
	}
}

func TestAdaptivePolicy_Reset(t *testing.T) {
	policy := NewAdaptivePolicy()

	metrics := types.LatencyMetrics{
		CaptureLatency:  50 * time.Millisecond,
		PlaybackLatency: 40 * time.Millisecond,
	}

	policy.Evaluate(metrics, 0, 0.5)

	if policy.GetActionsApplied() == 0 {
		t.Error("Expected non-zero actions before reset")
	}

	policy.Reset()

	if policy.GetActionsApplied() != 0 {
		t.Error("Expected 0 actions after reset")
	}
}

func TestAdaptivePolicy_GetStats(t *testing.T) {
	policy := NewAdaptivePolicy()

	metrics := types.LatencyMetrics{
		CaptureLatency:  50 * time.Millisecond,
		PlaybackLatency: 40 * time.Millisecond,
	}

	policy.Evaluate(metrics, 0, 0.5)

	stats := policy.GetStats()

	if stats.LatencyThreshold != 80*time.Millisecond {
		t.Errorf("Expected latency threshold 80ms, got %v", stats.LatencyThreshold)
	}
	if stats.BufferAdjustmentStep != 2 {
		t.Errorf("Expected buffer step 2, got %d", stats.BufferAdjustmentStep)
	}
	if stats.CPUThreshold != 0.8 {
		t.Errorf("Expected CPU threshold 0.8, got %f", stats.CPUThreshold)
	}
	if stats.UnderrunThreshold != 5 {
		t.Errorf("Expected underrun threshold 5, got %d", stats.UnderrunThreshold)
	}
	if stats.ActionsApplied == 0 {
		t.Error("Expected non-zero actions applied")
	}
}

func TestActionType_String(t *testing.T) {
	tests := []struct {
		actionType ActionType
		expected   string
	}{
		{ReduceBuffer, "ReduceBuffer"},
		{IncreaseBuffer, "IncreaseBuffer"},
		{SwitchToExclusiveMode, "SwitchToExclusiveMode"},
		{SwitchToSharedMode, "SwitchToSharedMode"},
		{ApplyDriftCompensation, "ApplyDriftCompensation"},
	}

	for _, tt := range tests {
		result := tt.actionType.String()
		if result != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, result)
		}
	}
}

func TestAdaptivePolicy_ConcurrentAccess(t *testing.T) {
	policy := NewAdaptivePolicy()

	done := make(chan bool)

	// Multiple goroutines evaluating policy
	for i := 0; i < 10; i++ {
		go func() {
			metrics := types.LatencyMetrics{
				CaptureLatency:  20 * time.Millisecond,
				PlaybackLatency: 30 * time.Millisecond,
			}

			for j := 0; j < 100; j++ {
				policy.Evaluate(metrics, 0, 0.5)
				policy.GetStats()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic or deadlock
	t.Logf("Concurrent test completed with %d actions applied", policy.GetActionsApplied())
}
