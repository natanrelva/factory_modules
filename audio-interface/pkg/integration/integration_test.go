package integration

import (
	"github.com/dubbing-system/audio-interface/pkg/types"
	"testing"
	"time"
)

// ASR Interface Tests

func TestNewASRInterface(t *testing.T) {
	asr := NewASRInterface()
	if asr == nil {
		t.Fatal("NewASRInterface returned nil")
	}
}

func TestASRInterface_StartStop(t *testing.T) {
	asr := NewASRInterface()

	if asr.IsRunning() {
		t.Error("Should not be running initially")
	}

	err := asr.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !asr.IsRunning() {
		t.Error("Should be running after Start")
	}

	err = asr.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if asr.IsRunning() {
		t.Error("Should not be running after Stop")
	}
}

func TestASRInterface_SendFrame(t *testing.T) {
	asr := NewASRInterface()
	asr.Start()
	defer asr.Stop()

	frame := types.PCMFrame{
		Data:       make([]int16, 320),
		SampleRate: 16000,
		Channels:   1,
		Timestamp:  time.Now(),
		Duration:   20 * time.Millisecond,
	}

	err := asr.SendFrame(frame)
	if err != nil {
		t.Errorf("SendFrame failed: %v", err)
	}

	framesSent, _ := asr.GetStats()
	if framesSent != 1 {
		t.Errorf("Expected 1 frame sent, got %d", framesSent)
	}
}

func TestASRInterface_SendFrameNotRunning(t *testing.T) {
	asr := NewASRInterface()

	frame := types.PCMFrame{
		Data: make([]int16, 320),
	}

	err := asr.SendFrame(frame)
	if err == nil {
		t.Error("Expected error when sending frame while not running")
	}
}

func TestASRInterface_ReceiveResult(t *testing.T) {
	asr := NewASRInterface()
	asr.Start()
	defer asr.Stop()

	result := ASRResult{
		Text:       "olá mundo",
		Confidence: 0.95,
		Timestamp:  time.Now(),
		IsFinal:    true,
		Language:   "pt-BR",
	}

	err := asr.SendResult(result)
	if err != nil {
		t.Errorf("SendResult failed: %v", err)
	}

	select {
	case received := <-asr.ReceiveResult():
		if received.Text != result.Text {
			t.Errorf("Expected text '%s', got '%s'", result.Text, received.Text)
		}
		if received.Confidence != result.Confidence {
			t.Errorf("Expected confidence %f, got %f", result.Confidence, received.Confidence)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for result")
	}
}

func TestASRInterface_GetLatency(t *testing.T) {
	asr := NewASRInterface()

	latency := asr.GetLatency()
	if latency != 200*time.Millisecond {
		t.Errorf("Expected default latency 200ms, got %v", latency)
	}

	asr.SetLatency(150 * time.Millisecond)
	latency = asr.GetLatency()
	if latency != 150*time.Millisecond {
		t.Errorf("Expected latency 150ms, got %v", latency)
	}
}

func TestASRInterface_Reset(t *testing.T) {
	asr := NewASRInterface()
	asr.Start()
	defer asr.Stop()

	frame := types.PCMFrame{Data: make([]int16, 320)}
	asr.SendFrame(frame)

	framesSent, _ := asr.GetStats()
	if framesSent == 0 {
		t.Error("Expected non-zero frames sent before reset")
	}

	asr.Reset()

	framesSent, resultsRecv := asr.GetStats()
	if framesSent != 0 || resultsRecv != 0 {
		t.Errorf("Expected zero stats after reset, got frames=%d, results=%d", framesSent, resultsRecv)
	}
}

// TTS Interface Tests

func TestNewTTSInterface(t *testing.T) {
	tts := NewTTSInterface()
	if tts == nil {
		t.Fatal("NewTTSInterface returned nil")
	}
}

func TestTTSInterface_StartStop(t *testing.T) {
	tts := NewTTSInterface()

	if tts.IsRunning() {
		t.Error("Should not be running initially")
	}

	err := tts.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if !tts.IsRunning() {
		t.Error("Should be running after Start")
	}

	err = tts.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if tts.IsRunning() {
		t.Error("Should not be running after Stop")
	}
}

func TestTTSInterface_SendText(t *testing.T) {
	tts := NewTTSInterface()
	tts.Start()
	defer tts.Stop()

	text := "hello world"
	err := tts.SendText(text)
	if err != nil {
		t.Errorf("SendText failed: %v", err)
	}

	textsSent, _ := tts.GetStats()
	if textsSent != 1 {
		t.Errorf("Expected 1 text sent, got %d", textsSent)
	}
}

func TestTTSInterface_SendTextNotRunning(t *testing.T) {
	tts := NewTTSInterface()

	err := tts.SendText("test")
	if err == nil {
		t.Error("Expected error when sending text while not running")
	}
}

func TestTTSInterface_ReceiveFrame(t *testing.T) {
	tts := NewTTSInterface()
	tts.Start()
	defer tts.Stop()

	frame := types.PCMFrame{
		Data:       make([]int16, 320),
		SampleRate: 16000,
		Channels:   1,
		Timestamp:  time.Now(),
		Duration:   20 * time.Millisecond,
	}

	err := tts.SendFrame(frame)
	if err != nil {
		t.Errorf("SendFrame failed: %v", err)
	}

	select {
	case received := <-tts.ReceiveFrame():
		if len(received.Data) != len(frame.Data) {
			t.Errorf("Expected frame size %d, got %d", len(frame.Data), len(received.Data))
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for frame")
	}
}

func TestTTSInterface_GetLatency(t *testing.T) {
	tts := NewTTSInterface()

	latency := tts.GetLatency()
	if latency != 200*time.Millisecond {
		t.Errorf("Expected default latency 200ms, got %v", latency)
	}

	tts.SetLatency(180 * time.Millisecond)
	latency = tts.GetLatency()
	if latency != 180*time.Millisecond {
		t.Errorf("Expected latency 180ms, got %v", latency)
	}
}

func TestTTSInterface_SendTextWithMetadata(t *testing.T) {
	tts := NewTTSInterface()
	tts.Start()
	defer tts.Stop()

	metadata := TTSMetadata{
		Language:     "en-US",
		VoiceID:      "voice-001",
		SpeakerEmbed: []float32{0.1, 0.2, 0.3},
		ProsodyHints: ProsodyInfo{
			RelativeDuration: 1.0,
			EmphasisLevel:    1,
			PauseAfter:       100 * time.Millisecond,
		},
	}

	err := tts.SendTextWithMetadata("hello", metadata)
	if err != nil {
		t.Errorf("SendTextWithMetadata failed: %v", err)
	}
}

func TestTTSInterface_Reset(t *testing.T) {
	tts := NewTTSInterface()
	tts.Start()
	defer tts.Stop()

	tts.SendText("test")

	textsSent, _ := tts.GetStats()
	if textsSent == 0 {
		t.Error("Expected non-zero texts sent before reset")
	}

	tts.Reset()

	textsSent, framesRecv := tts.GetStats()
	if textsSent != 0 || framesRecv != 0 {
		t.Errorf("Expected zero stats after reset, got texts=%d, frames=%d", textsSent, framesRecv)
	}
}

// Integration Tests

func TestASRTTSIntegration(t *testing.T) {
	asr := NewASRInterface()
	tts := NewTTSInterface()

	asr.Start()
	tts.Start()
	defer asr.Stop()
	defer tts.Stop()

	// Simulate ASR → TTS flow
	// 1. Send audio frame to ASR
	frame := types.PCMFrame{
		Data:       make([]int16, 320),
		SampleRate: 16000,
		Channels:   1,
		Timestamp:  time.Now(),
		Duration:   20 * time.Millisecond,
	}

	err := asr.SendFrame(frame)
	if err != nil {
		t.Fatalf("ASR SendFrame failed: %v", err)
	}

	// 2. Simulate ASR recognition result
	result := ASRResult{
		Text:       "olá",
		Confidence: 0.95,
		Timestamp:  time.Now(),
		IsFinal:    true,
		Language:   "pt-BR",
	}

	err = asr.SendResult(result)
	if err != nil {
		t.Fatalf("ASR SendResult failed: %v", err)
	}

	// 3. Receive ASR result
	select {
	case asrResult := <-asr.ReceiveResult():
		// 4. Send translated text to TTS
		translatedText := "hello" // Simulated translation
		err = tts.SendText(translatedText)
		if err != nil {
			t.Fatalf("TTS SendText failed: %v", err)
		}

		t.Logf("ASR recognized: %s (confidence: %.2f)", asrResult.Text, asrResult.Confidence)
		t.Logf("TTS synthesizing: %s", translatedText)

	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for ASR result")
	}

	// 5. Simulate TTS synthesis
	synthesizedFrame := types.PCMFrame{
		Data:       make([]int16, 320),
		SampleRate: 16000,
		Channels:   1,
		Timestamp:  time.Now(),
		Duration:   20 * time.Millisecond,
	}

	err = tts.SendFrame(synthesizedFrame)
	if err != nil {
		t.Fatalf("TTS SendFrame failed: %v", err)
	}

	// 6. Receive synthesized frame
	select {
	case <-tts.ReceiveFrame():
		t.Log("Received synthesized audio frame")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for TTS frame")
	}
}
