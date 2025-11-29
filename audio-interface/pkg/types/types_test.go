package types

import (
	"testing"
	"time"
)

func TestPCMFrame_Creation(t *testing.T) {
	frame := PCMFrame{
		Data:       make([]int16, 320), // 20ms at 16kHz
		SampleRate: 16000,
		Channels:   1,
		Timestamp:  time.Now(),
		Duration:   20 * time.Millisecond,
		IsSpeech:   true,
	}

	if len(frame.Data) != 320 {
		t.Errorf("Expected 320 samples, got %d", len(frame.Data))
	}

	if frame.SampleRate != 16000 {
		t.Errorf("Expected sample rate 16000, got %d", frame.SampleRate)
	}

	if frame.Duration != 20*time.Millisecond {
		t.Errorf("Expected duration 20ms, got %v", frame.Duration)
	}
}

func TestAudioConfig_Validation(t *testing.T) {
	config := AudioConfig{
		DeviceID:   "",
		SampleRate: 48000,
		Channels:   2,
		FrameSize:  960, // 20ms at 48kHz
		BufferSize: 10,
	}

	if config.SampleRate != 48000 {
		t.Errorf("Expected sample rate 48000, got %d", config.SampleRate)
	}

	if config.Channels != 2 {
		t.Errorf("Expected 2 channels, got %d", config.Channels)
	}
}

func TestLatencyMetrics_Creation(t *testing.T) {
	metrics := LatencyMetrics{
		CaptureLatency:  25 * time.Millisecond,
		PlaybackLatency: 35 * time.Millisecond,
		BufferFillLevel: 0.6,
		DroppedFrames:   0,
		Underruns:       0,
		Overruns:        0,
		Timestamp:       time.Now(),
	}

	if metrics.CaptureLatency != 25*time.Millisecond {
		t.Errorf("Expected capture latency 25ms, got %v", metrics.CaptureLatency)
	}

	if metrics.BufferFillLevel != 0.6 {
		t.Errorf("Expected buffer fill level 0.6, got %f", metrics.BufferFillLevel)
	}
}

func TestWASAPIMode_Values(t *testing.T) {
	if Shared != 0 {
		t.Errorf("Expected Shared mode to be 0, got %d", Shared)
	}

	if Exclusive != 1 {
		t.Errorf("Expected Exclusive mode to be 1, got %d", Exclusive)
	}
}

func TestErrorInfo_Creation(t *testing.T) {
	errInfo := ErrorInfo{
		Module:    "TestModule",
		Operation: "TestOperation",
		Timestamp: time.Now(),
		Context:   "Test context",
	}

	if errInfo.Module != "TestModule" {
		t.Errorf("Expected module 'TestModule', got '%s'", errInfo.Module)
	}

	if errInfo.Operation != "TestOperation" {
		t.Errorf("Expected operation 'TestOperation', got '%s'", errInfo.Operation)
	}
}
