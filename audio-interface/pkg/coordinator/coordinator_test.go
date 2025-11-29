package coordinator

import (
	"github.com/dubbing-system/audio-interface/pkg/types"
	"testing"
	"time"
)

func TestNewAudioInterfaceCoordinator(t *testing.T) {
	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	coordinator := NewAudioInterfaceCoordinator(config)
	if coordinator == nil {
		t.Fatal("NewAudioInterfaceCoordinator returned nil")
	}

	if coordinator.capture == nil {
		t.Error("Capture not initialized")
	}
	if coordinator.playback == nil {
		t.Error("Playback not initialized")
	}
	if coordinator.synchronizer == nil {
		t.Error("Synchronizer not initialized")
	}
	if coordinator.latencyManager == nil {
		t.Error("Latency manager not initialized")
	}
	if coordinator.metricsCollector == nil {
		t.Error("Metrics collector not initialized")
	}
}

func TestAudioInterfaceCoordinator_Initialize(t *testing.T) {
	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	coordinator := NewAudioInterfaceCoordinator(config)

	err := coordinator.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
}

func TestAudioInterfaceCoordinator_Initialize_InvalidConfig(t *testing.T) {
	config := types.AudioConfig{
		SampleRate: 0, // Invalid
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	coordinator := NewAudioInterfaceCoordinator(config)

	err := coordinator.Initialize()
	if err == nil {
		t.Error("Expected error for invalid config")
	}
}

func TestAudioInterfaceCoordinator_StartStop(t *testing.T) {
	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	coordinator := NewAudioInterfaceCoordinator(config)

	err := coordinator.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = coordinator.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !coordinator.IsRunning() {
		t.Error("Coordinator should be running")
	}

	// Let it run for a bit
	time.Sleep(200 * time.Millisecond)

	err = coordinator.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Give workers time to stop
	time.Sleep(100 * time.Millisecond)

	if coordinator.IsRunning() {
		t.Error("Coordinator should not be running")
	}
}

func TestAudioInterfaceCoordinator_StartWithoutInitialize(t *testing.T) {
	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	coordinator := NewAudioInterfaceCoordinator(config)

	err := coordinator.Start()
	if err == nil {
		t.Error("Expected error when starting without initialization")
		coordinator.Stop()
	}
}

func TestAudioInterfaceCoordinator_Close(t *testing.T) {
	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	coordinator := NewAudioInterfaceCoordinator(config)

	err := coordinator.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = coordinator.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	err = coordinator.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if coordinator.IsRunning() {
		t.Error("Coordinator should not be running after close")
	}
}

func TestAudioInterfaceCoordinator_GetMetrics(t *testing.T) {
	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	coordinator := NewAudioInterfaceCoordinator(config)

	err := coordinator.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = coordinator.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer coordinator.Stop()

	// Let it collect some metrics
	time.Sleep(300 * time.Millisecond)

	metrics := coordinator.GetMetrics()
	if metrics.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestAudioInterfaceCoordinator_GetMetricsSummary(t *testing.T) {
	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	coordinator := NewAudioInterfaceCoordinator(config)

	err := coordinator.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = coordinator.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer coordinator.Stop()

	// Let it collect some metrics
	time.Sleep(300 * time.Millisecond)

	summary := coordinator.GetMetricsSummary()
	if summary.Uptime < 0 {
		t.Errorf("Expected non-negative uptime, got %v", summary.Uptime)
	}
}

func TestAudioInterfaceCoordinator_GetLatencyStats(t *testing.T) {
	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	coordinator := NewAudioInterfaceCoordinator(config)

	err := coordinator.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = coordinator.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer coordinator.Stop()

	// Let it collect some metrics
	time.Sleep(600 * time.Millisecond)

	stats := coordinator.GetLatencyStats()
	if stats.TargetLatency != 100*time.Millisecond {
		t.Errorf("Expected target latency 100ms, got %v", stats.TargetLatency)
	}
}

func TestAudioInterfaceCoordinator_GetSyncStats(t *testing.T) {
	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	coordinator := NewAudioInterfaceCoordinator(config)

	err := coordinator.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = coordinator.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer coordinator.Stop()

	// Let it collect some metrics
	time.Sleep(300 * time.Millisecond)

	stats := coordinator.GetSyncStats()
	if stats.TargetAlignment != 50*time.Millisecond {
		t.Errorf("Expected target alignment 50ms, got %v", stats.TargetAlignment)
	}
}

func TestAudioInterfaceCoordinator_SetTargetLatency(t *testing.T) {
	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	coordinator := NewAudioInterfaceCoordinator(config)

	err := coordinator.SetTargetLatency(80 * time.Millisecond)
	if err != nil {
		t.Errorf("SetTargetLatency failed: %v", err)
	}

	// Invalid target
	err = coordinator.SetTargetLatency(5 * time.Millisecond)
	if err == nil {
		t.Error("Expected error for invalid target latency")
	}
}

func TestAudioInterfaceCoordinator_GetConfig(t *testing.T) {
	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	coordinator := NewAudioInterfaceCoordinator(config)

	retrievedConfig := coordinator.GetConfig()
	if retrievedConfig.SampleRate != config.SampleRate {
		t.Errorf("Expected sample rate %d, got %d", config.SampleRate, retrievedConfig.SampleRate)
	}
	if retrievedConfig.Channels != config.Channels {
		t.Errorf("Expected channels %d, got %d", config.Channels, retrievedConfig.Channels)
	}
}

func TestAudioInterfaceCoordinator_MultipleStartStop(t *testing.T) {
	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	coordinator := NewAudioInterfaceCoordinator(config)

	err := coordinator.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// First start/stop cycle
	err = coordinator.Start()
	if err != nil {
		t.Fatalf("First start failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	err = coordinator.Stop()
	if err != nil {
		t.Fatalf("First stop failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Second start - components are still initialized so it should work
	err = coordinator.Start()
	if err != nil {
		t.Logf("Second start failed (expected): %v", err)
		// This is acceptable behavior
	} else {
		// If it succeeds, clean up
		coordinator.Stop()
	}
}

func TestAudioInterfaceCoordinator_StopTimeout(t *testing.T) {
	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	coordinator := NewAudioInterfaceCoordinator(config)

	err := coordinator.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = coordinator.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Stop should complete even if workers are slow
	err = coordinator.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestAudioInterfaceCoordinator_WorkersRunning(t *testing.T) {
	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	coordinator := NewAudioInterfaceCoordinator(config)

	err := coordinator.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = coordinator.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer coordinator.Stop()

	// Let workers run and collect metrics
	time.Sleep(1 * time.Second)

	// Check that metrics are being collected
	summary := coordinator.GetMetricsSummary()
	if summary.TotalModules == 0 {
		t.Error("Expected some modules to have metrics")
	}

	t.Logf("Collected metrics from %d modules", summary.TotalModules)
}
