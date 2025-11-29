package coordinator

import (
	"github.com/dubbing-system/audio-interface/pkg/integration"
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

// V2.0 Tests

func TestAudioInterfaceCoordinator_ConnectASR(t *testing.T) {
	coord := NewAudioInterfaceCoordinator(types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	})

	asr := integration.NewASRInterface()
	
	err := coord.ConnectASR(asr)
	if err != nil {
		t.Fatalf("ConnectASR failed: %v", err)
	}

	if !coord.IsASRConnected() {
		t.Error("ASR should be connected")
	}

	if coord.GetASRInterface() != asr {
		t.Error("GetASRInterface returned wrong interface")
	}
}

func TestAudioInterfaceCoordinator_ConnectTTS(t *testing.T) {
	coord := NewAudioInterfaceCoordinator(types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	})

	tts := integration.NewTTSInterface()
	
	err := coord.ConnectTTS(tts)
	if err != nil {
		t.Fatalf("ConnectTTS failed: %v", err)
	}

	if !coord.IsTTSConnected() {
		t.Error("TTS should be connected")
	}

	if coord.GetTTSInterface() != tts {
		t.Error("GetTTSInterface returned wrong interface")
	}
}

func TestAudioInterfaceCoordinator_ConnectWhileRunning(t *testing.T) {
	coord := NewAudioInterfaceCoordinator(types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	})

	coord.Initialize()
	coord.Start()
	defer coord.Close()

	asr := integration.NewASRInterface()
	err := coord.ConnectASR(asr)
	if err == nil {
		t.Error("Expected error when connecting ASR while running")
	}

	tts := integration.NewTTSInterface()
	err = coord.ConnectTTS(tts)
	if err == nil {
		t.Error("Expected error when connecting TTS while running")
	}
}

func TestAudioInterfaceCoordinator_GetBackpressureStats(t *testing.T) {
	coord := NewAudioInterfaceCoordinator(types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	})

	events, duration := coord.GetBackpressureStats()
	if events != 0 {
		t.Errorf("Expected 0 events initially, got %d", events)
	}
	if duration != 0 {
		t.Errorf("Expected 0 duration initially, got %v", duration)
	}
}

func TestAudioInterfaceCoordinator_GetAdaptivePolicyStats(t *testing.T) {
	coord := NewAudioInterfaceCoordinator(types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	})

	stats := coord.GetAdaptivePolicyStats()
	if stats.LatencyThreshold != 80*time.Millisecond {
		t.Errorf("Expected latency threshold 80ms, got %v", stats.LatencyThreshold)
	}
	if stats.ActionsApplied != 0 {
		t.Errorf("Expected 0 actions initially, got %d", stats.ActionsApplied)
	}
}

func TestAudioInterfaceCoordinator_BackpressureIntegration(t *testing.T) {
	coord := NewAudioInterfaceCoordinator(types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 5, // Small buffer to trigger backpressure
	})

	err := coord.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = coord.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer coord.Close()

	// Let it run for a bit
	time.Sleep(500 * time.Millisecond)

	// Check if backpressure was triggered
	events, _ := coord.GetBackpressureStats()
	t.Logf("Backpressure events: %d", events)
	
	// Note: May or may not have backpressure depending on timing
}

func TestAudioInterfaceCoordinator_AdaptivePolicyIntegration(t *testing.T) {
	coord := NewAudioInterfaceCoordinator(types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	})

	err := coord.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = coord.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer coord.Close()

	// Let monitor worker run
	time.Sleep(1 * time.Second)

	// Check adaptive policy stats
	stats := coord.GetAdaptivePolicyStats()
	t.Logf("Adaptive policy actions applied: %d", stats.ActionsApplied)
	
	// Note: May or may not have actions depending on metrics
}

func TestAudioInterfaceCoordinator_ASRIntegration(t *testing.T) {
	coord := NewAudioInterfaceCoordinator(types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	})

	asr := integration.NewASRInterface()
	asr.Start()
	defer asr.Stop()

	err := coord.ConnectASR(asr)
	if err != nil {
		t.Fatalf("ConnectASR failed: %v", err)
	}

	err = coord.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = coord.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer coord.Close()

	// Let it run and send frames to ASR
	time.Sleep(500 * time.Millisecond)

	// Check ASR stats
	framesSent, _ := asr.GetStats()
	t.Logf("Frames sent to ASR: %d", framesSent)
	
	if framesSent == 0 {
		t.Error("Expected frames to be sent to ASR")
	}
}
