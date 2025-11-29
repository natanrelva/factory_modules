package capture

import (
	"github.com/dubbing-system/audio-interface/pkg/types"
	"testing"
	"time"
)

func TestNewWindowsAudioCapture(t *testing.T) {
	capture := NewWindowsAudioCapture()
	if capture == nil {
		t.Fatal("NewWindowsAudioCapture returned nil")
	}
	if capture.frameChannel == nil {
		t.Error("Frame channel not initialized")
	}
}

func TestWindowsAudioCapture_Initialize(t *testing.T) {
	capture := NewWindowsAudioCapture()

	config := types.AudioConfig{
		DeviceID:   "",
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320, // 20ms at 16kHz
		BufferSize: 10,
	}

	err := capture.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if !capture.initialized {
		t.Error("Capture should be initialized")
	}
}

func TestWindowsAudioCapture_Initialize_InvalidConfig(t *testing.T) {
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
			capture := NewWindowsAudioCapture()
			err := capture.Initialize(tt.config)
			if err == nil {
				t.Error("Expected error for invalid config, got nil")
			}
		})
	}
}

func TestWindowsAudioCapture_StartStop(t *testing.T) {
	capture := NewWindowsAudioCapture()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	err := capture.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = capture.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !capture.running {
		t.Error("Capture should be running")
	}

	err = capture.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	// Give goroutine time to stop
	time.Sleep(50 * time.Millisecond)

	if capture.running {
		t.Error("Capture should not be running")
	}
}

func TestWindowsAudioCapture_StartWithoutInitialize(t *testing.T) {
	capture := NewWindowsAudioCapture()

	err := capture.Start()
	if err == nil {
		t.Error("Expected error when starting without initialization")
	}
}

func TestWindowsAudioCapture_FrameGeneration(t *testing.T) {
	capture := NewWindowsAudioCapture()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320, // 20ms at 16kHz
		BufferSize: 10,
	}

	err := capture.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = capture.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer capture.Stop()

	// Wait for frames
	timeout := time.After(200 * time.Millisecond)
	frameCount := 0

	for frameCount < 5 {
		select {
		case frame := <-capture.GetFrameChannel():
			frameCount++

			// Validate frame
			if frame.SampleRate != config.SampleRate {
				t.Errorf("Expected sample rate %d, got %d", config.SampleRate, frame.SampleRate)
			}
			if frame.Channels != config.Channels {
				t.Errorf("Expected %d channels, got %d", config.Channels, frame.Channels)
			}
			if len(frame.Data) != config.FrameSize*config.Channels {
				t.Errorf("Expected data length %d, got %d", config.FrameSize*config.Channels, len(frame.Data))
			}

			// Check frame duration (should be 10-20ms)
			durationMs := frame.Duration.Milliseconds()
			if durationMs < 10 || durationMs > 20 {
				t.Errorf("Frame duration %dms outside expected range [10-20ms]", durationMs)
			}

		case <-timeout:
			t.Fatalf("Timeout waiting for frames, got %d frames", frameCount)
		}
	}

	if frameCount < 5 {
		t.Errorf("Expected at least 5 frames, got %d", frameCount)
	}
}

func TestWindowsAudioCapture_CaptureLatency(t *testing.T) {
	capture := NewWindowsAudioCapture()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	err := capture.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = capture.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer capture.Stop()

	// Wait for some frames to be captured
	time.Sleep(100 * time.Millisecond)

	latency := capture.GetCaptureLatency()
	
	// Latency should be reasonable (< 50ms for this mock implementation)
	if latency > 50*time.Millisecond {
		t.Errorf("Capture latency too high: %v", latency)
	}
}

func TestWindowsAudioCapture_Close(t *testing.T) {
	capture := NewWindowsAudioCapture()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	err := capture.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = capture.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	err = capture.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if capture.initialized {
		t.Error("Capture should not be initialized after close")
	}
}

func TestWindowsAudioCapture_CalculateFrameDuration(t *testing.T) {
	tests := []struct {
		name           string
		sampleRate     int
		frameSize      int
		expectedMs     int64
		toleranceMs    int64
	}{
		{
			name:        "16kHz 20ms",
			sampleRate:  16000,
			frameSize:   320,
			expectedMs:  20,
			toleranceMs: 1,
		},
		{
			name:        "48kHz 10ms",
			sampleRate:  48000,
			frameSize:   480,
			expectedMs:  10,
			toleranceMs: 1,
		},
		{
			name:        "16kHz 10ms",
			sampleRate:  16000,
			frameSize:   160,
			expectedMs:  10,
			toleranceMs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := NewWindowsAudioCapture()
			config := types.AudioConfig{
				SampleRate: tt.sampleRate,
				Channels:   1,
				FrameSize:  tt.frameSize,
				BufferSize: 10,
			}

			capture.Initialize(config)
			duration := capture.calculateFrameDuration()
			durationMs := duration.Milliseconds()

			diff := durationMs - tt.expectedMs
			if diff < 0 {
				diff = -diff
			}

			if diff > tt.toleranceMs {
				t.Errorf("Expected duration ~%dms, got %dms", tt.expectedMs, durationMs)
			}
		})
	}
}

func TestWindowsAudioCapture_Stats(t *testing.T) {
	capture := NewWindowsAudioCapture()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	err := capture.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	errors, fillLevel := capture.GetStats()
	if errors != 0 {
		t.Errorf("Expected 0 errors initially, got %d", errors)
	}
	if fillLevel != 0.0 {
		t.Errorf("Expected 0.0 fill level initially, got %f", fillLevel)
	}

	capture.ResetStats()
	errors, _ = capture.GetStats()
	if errors != 0 {
		t.Errorf("Expected 0 errors after reset, got %d", errors)
	}
}

func TestWindowsAudioCapture_ContinuousCapture(t *testing.T) {
	capture := NewWindowsAudioCapture()

	config := types.AudioConfig{
		SampleRate: 16000,
		Channels:   1,
		FrameSize:  320,
		BufferSize: 10,
	}

	err := capture.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	err = capture.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer capture.Stop()

	// Capture for 500ms and verify continuity
	startTime := time.Now()
	frameCount := 0
	var lastTimestamp time.Time

	for time.Since(startTime) < 500*time.Millisecond {
		select {
		case frame := <-capture.GetFrameChannel():
			frameCount++
			
			// Check timestamp progression
			if !lastTimestamp.IsZero() {
				timeDiff := frame.Timestamp.Sub(lastTimestamp)
				// Timestamps should progress (allow some jitter)
				if timeDiff < 0 || timeDiff > 50*time.Millisecond {
					t.Errorf("Unexpected timestamp gap: %v", timeDiff)
				}
			}
			lastTimestamp = frame.Timestamp

		case <-time.After(100 * time.Millisecond):
			// No frame received in 100ms - might be an issue
		}
	}

	// Should have captured some frames (at least 5 in 500ms)
	// Allow wide range due to timing variations and goroutine scheduling
	if frameCount < 5 {
		t.Errorf("Expected at least 5 frames in 500ms, got %d", frameCount)
	}
	
	t.Logf("Captured %d frames in 500ms", frameCount)
}
