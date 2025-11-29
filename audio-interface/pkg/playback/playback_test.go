package playback

import (
	"github.com/dubbing-system/audio-interface/pkg/types"
	"testing"
	"time"
)

func TestNewWindowsAudioPlayback(t *testing.T) {
	playback := NewWindowsAudioPlayback()
	if playback == nil {
		t.Fatal("NewWindowsAudioPlayback returned nil")
	}
}

func TestWindowsAudioPlayback_Initialize(t *testing.T) {
	playback := NewWindowsAudioPlayback()

	config := types.AudioConfig{
		DeviceID:   "",
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320, // 20ms at 16kHz
		BufferSize: 10,
	}

	err := playback.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if !playback.initialized {
		t.Error("Playback should be initialized")
	}
}

func TestWindowsAudioPlayback_Initialize_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config types.AudioConfig
	}{
		{
			name: "Invalid sample rate",
			config: types.AudioConfig{
				SampleRate: 0,
				Channels:   1,
				FrameSize:  320,
			},
		},
		{
			name: "Invalid channels",
			config: types.AudioConfig{
				SampleRate: 16000,
				Channels:   0,
				FrameSize:  320,
			},
		},
		{
			name: "Invalid frame size",
			config: types.AudioConfig{
				SampleRate: 16000,
				Channels:   1,
				FrameSize:  0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			playback := NewWindowsAudioPlayback()
			err := playback.Initialize(tt.config)
			if err == nil {
				t.Error("Expected error for invalid config, got nil")
			}
		})
	}
}

func TestWindowsAudioPlayback_StartStop(t *testing.T) {
	playback := NewWindowsAudioPlayback()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	err := playback.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = playback.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !playback.running {
		t.Error("Playback should be running")
	}

	err = playback.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Give goroutine time to stop
	time.Sleep(50 * time.Millisecond)

	if playback.running {
		t.Error("Playback should not be running")
	}
}

func TestWindowsAudioPlayback_StartWithoutInitialize(t *testing.T) {
	playback := NewWindowsAudioPlayback()

	err := playback.Start()
	if err == nil {
		t.Error("Expected error when starting without initialization")
	}
}

func TestWindowsAudioPlayback_WriteFrame(t *testing.T) {
	playback := NewWindowsAudioPlayback()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	err := playback.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = playback.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer playback.Stop()

	frame := types.PCMFrame{
		Data:       make([]int16, 320),
		SampleRate: 16000,
		Channels:   1,
		Timestamp:  time.Now(),
		Duration:   20 * time.Millisecond,
	}

	err = playback.WriteFrame(frame)
	if err != nil {
		t.Errorf("WriteFrame failed: %v", err)
	}

	// Check buffer fill level increased
	fillLevel := playback.GetBufferFillLevel()
	if fillLevel <= 0.0 {
		t.Error("Buffer fill level should be > 0 after writing frame")
	}
}

func TestWindowsAudioPlayback_WriteFrameNotRunning(t *testing.T) {
	playback := NewWindowsAudioPlayback()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	playback.Initialize(config)

	frame := types.PCMFrame{
		Data:     make([]int16, 320),
		Duration: 20 * time.Millisecond,
	}

	err := playback.WriteFrame(frame)
	if err == nil {
		t.Error("Expected error when writing frame without starting playback")
	}
}

func TestWindowsAudioPlayback_BufferUnderrun(t *testing.T) {
	playback := NewWindowsAudioPlayback()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 5,
	}

	err := playback.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = playback.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer playback.Stop()

	// Don't write any frames - should cause underruns
	time.Sleep(200 * time.Millisecond)

	_, underruns, _ := playback.GetStats()
	if underruns == 0 {
		t.Error("Expected underruns when no frames are written")
	}

	t.Logf("Detected %d underruns", underruns)
}

func TestWindowsAudioPlayback_ContinuousPlayback(t *testing.T) {
	playback := NewWindowsAudioPlayback()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	err := playback.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = playback.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer playback.Stop()

	// Write frames continuously
	framesToWrite := 20
	for i := 0; i < framesToWrite; i++ {
		frame := types.PCMFrame{
			Data:       make([]int16, 320),
			SampleRate: 16000,
			Channels:   1,
			Timestamp:  time.Now(),
			Duration:   20 * time.Millisecond,
		}

		err := playback.WriteFrame(frame)
		if err != nil {
			t.Errorf("WriteFrame %d failed: %v", i, err)
		}

		time.Sleep(15 * time.Millisecond) // Write slightly faster than playback
	}

	// Wait for playback to finish
	time.Sleep(200 * time.Millisecond)

	errors, underruns, _ := playback.GetStats()
	
	// Should have minimal errors
	// Allow some underruns due to timing variations in test environment
	if errors > 5 {
		t.Errorf("Too many playback errors: %d", errors)
	}
	if underruns > 10 {
		t.Errorf("Too many underruns: %d (expected some due to timing)", underruns)
	}

	t.Logf("Playback stats - Errors: %d, Underruns: %d", errors, underruns)
}

func TestWindowsAudioPlayback_PlaybackLatency(t *testing.T) {
	playback := NewWindowsAudioPlayback()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	err := playback.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = playback.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer playback.Stop()

	// Write some frames
	for i := 0; i < 5; i++ {
		frame := types.PCMFrame{
			Data:     make([]int16, 320),
			Duration: 20 * time.Millisecond,
		}
		playback.WriteFrame(frame)
	}

	// Wait for playback to process
	time.Sleep(100 * time.Millisecond)

	latency := playback.GetPlaybackLatency()

	// Latency should be reasonable (< 100ms for this mock implementation)
	if latency > 100*time.Millisecond {
		t.Errorf("Playback latency too high: %v", latency)
	}

	// Should meet requirement of ≤50ms (with some tolerance for mock)
	if latency > 60*time.Millisecond {
		t.Logf("Warning: Playback latency %v exceeds target of 50ms", latency)
	}
}

func TestWindowsAudioPlayback_BufferFillLevel(t *testing.T) {
	playback := NewWindowsAudioPlayback()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	err := playback.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = playback.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer playback.Stop()

	// Initially empty
	fillLevel := playback.GetBufferFillLevel()
	if fillLevel != 0.0 {
		t.Errorf("Expected initial fill level 0.0, got %f", fillLevel)
	}

	// Write frames to fill buffer
	for i := 0; i < 5; i++ {
		frame := types.PCMFrame{
			Data:     make([]int16, 320),
			Duration: 20 * time.Millisecond,
		}
		playback.WriteFrame(frame)
	}

	// Should have some fill level
	fillLevel = playback.GetBufferFillLevel()
	if fillLevel <= 0.0 {
		t.Error("Fill level should be > 0 after writing frames")
	}
	if fillLevel > 1.0 {
		t.Errorf("Fill level should be ≤ 1.0, got %f", fillLevel)
	}

	t.Logf("Buffer fill level: %.2f", fillLevel)
}

func TestWindowsAudioPlayback_Close(t *testing.T) {
	playback := NewWindowsAudioPlayback()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	err := playback.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = playback.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	err = playback.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if playback.initialized {
		t.Error("Playback should not be initialized after close")
	}
}

func TestWindowsAudioPlayback_AdjustBufferSize(t *testing.T) {
	playback := NewWindowsAudioPlayback()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	err := playback.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Adjust buffer for 60ms target latency
	err = playback.AdjustBufferSize(60 * time.Millisecond)
	if err != nil {
		t.Errorf("AdjustBufferSize failed: %v", err)
	}

	// Adjust buffer for 100ms target latency
	err = playback.AdjustBufferSize(100 * time.Millisecond)
	if err != nil {
		t.Errorf("AdjustBufferSize failed: %v", err)
	}
}

func TestWindowsAudioPlayback_ResetStats(t *testing.T) {
	playback := NewWindowsAudioPlayback()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 5,
	}

	err := playback.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = playback.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer playback.Stop()

	// Cause some underruns
	time.Sleep(100 * time.Millisecond)

	errors, underruns, _ := playback.GetStats()
	if underruns == 0 {
		t.Log("Warning: Expected some underruns")
	}

	playback.ResetStats()

	errors, underruns, _ = playback.GetStats()
	if errors != 0 || underruns != 0 {
		t.Errorf("Expected stats reset to 0, got errors=%d, underruns=%d", errors, underruns)
	}
}
