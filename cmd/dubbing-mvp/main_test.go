package main

import (
	"testing"
	
	asrsimple "github.com/user/audio-dubbing-system/pkg/asr-simple"
	translationsimple "github.com/user/audio-dubbing-system/pkg/translation-simple"
	ttssimple "github.com/user/audio-dubbing-system/pkg/tts-simple"
)

func TestASRInitialization(t *testing.T) {
	config := asrsimple.Config{
		ModelPath:  "../../models/ggml-tiny.bin",
		Language:   "pt",
		SampleRate: 16000,
	}
	
	asr, err := asrsimple.NewSimpleASR(config)
	if err != nil {
		t.Fatalf("Failed to initialize ASR: %v", err)
	}
	defer asr.Close()
	
	// Test transcription with mock audio
	samples := make([]float32, 16000) // 1 second of audio
	for i := range samples {
		samples[i] = 0.1 // Some energy
	}
	
	text, err := asr.Transcribe(samples)
	if err != nil {
		t.Fatalf("Transcription failed: %v", err)
	}
	
	t.Logf("Transcribed text: %s", text)
}

func TestTranslatorInitialization(t *testing.T) {
	config := translationsimple.Config{
		SourceLang: "pt",
		TargetLang: "en",
		UseAPI:     false, // Use mock
	}
	
	translator, err := translationsimple.NewSimpleTranslator(config)
	if err != nil {
		t.Fatalf("Failed to initialize translator: %v", err)
	}
	defer translator.Close()
	
	// Test translation
	textPT := "olá"
	textEN, err := translator.Translate(textPT)
	if err != nil {
		t.Fatalf("Translation failed: %v", err)
	}
	
	if textEN == "" {
		t.Fatal("Translation returned empty string")
	}
	
	t.Logf("Translation: %s → %s", textPT, textEN)
}

func TestTTSInitialization(t *testing.T) {
	config := ttssimple.Config{
		Voice:      "en-us-female",
		SampleRate: 16000,
		Engine:     "mock",
	}
	
	tts, err := ttssimple.NewSimpleTTS(config)
	if err != nil {
		t.Fatalf("Failed to initialize TTS: %v", err)
	}
	defer tts.Close()
	
	// Test synthesis
	textEN := "hello world"
	audio, err := tts.Synthesize(textEN)
	if err != nil {
		t.Fatalf("Synthesis failed: %v", err)
	}
	
	if len(audio) == 0 {
		t.Fatal("Synthesis returned no audio")
	}
	
	t.Logf("Synthesized %d samples for: %s", len(audio), textEN)
}

func TestPipelineIntegration(t *testing.T) {
	// Initialize all modules
	asr, err := asrsimple.NewSimpleASR(asrsimple.Config{
		ModelPath:  "../../models/ggml-tiny.bin",
		Language:   "pt",
		SampleRate: 16000,
	})
	if err != nil {
		t.Fatalf("ASR init failed: %v", err)
	}
	defer asr.Close()
	
	translator, err := translationsimple.NewSimpleTranslator(translationsimple.Config{
		SourceLang: "pt",
		TargetLang: "en",
		UseAPI:     false,
	})
	if err != nil {
		t.Fatalf("Translator init failed: %v", err)
	}
	defer translator.Close()
	
	tts, err := ttssimple.NewSimpleTTS(ttssimple.Config{
		Voice:      "en-us-female",
		SampleRate: 16000,
		Engine:     "mock",
	})
	if err != nil {
		t.Fatalf("TTS init failed: %v", err)
	}
	defer tts.Close()
	
	// Test full pipeline
	// 1. Mock audio input
	audioInput := make([]float32, 16000)
	for i := range audioInput {
		audioInput[i] = 0.1
	}
	
	// 2. ASR
	textPT, err := asr.Transcribe(audioInput)
	if err != nil {
		t.Fatalf("ASR failed: %v", err)
	}
	t.Logf("ASR output: %s", textPT)
	
	// 3. Translation
	textEN, err := translator.Translate(textPT)
	if err != nil {
		t.Fatalf("Translation failed: %v", err)
	}
	t.Logf("Translation output: %s", textEN)
	
	// 4. TTS
	audioOutput, err := tts.Synthesize(textEN)
	if err != nil {
		t.Fatalf("TTS failed: %v", err)
	}
	t.Logf("TTS output: %d samples", len(audioOutput))
	
	// Verify pipeline completed
	if len(audioOutput) == 0 {
		t.Fatal("Pipeline produced no audio output")
	}
	
	t.Log("✓ Pipeline integration test passed")
}
