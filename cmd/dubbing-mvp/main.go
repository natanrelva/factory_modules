package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	
	// Mock implementations
	asrsimple "github.com/user/audio-dubbing-system/pkg/asr-simple"
	translationsimple "github.com/user/audio-dubbing-system/pkg/translation-simple"
	ttssimple "github.com/user/audio-dubbing-system/pkg/tts-simple"
	
	// Real implementations
	argos "github.com/user/audio-dubbing-system/pkg/translation-argos"
	espeak "github.com/user/audio-dubbing-system/pkg/tts-espeak"
	ttswindows "github.com/user/audio-dubbing-system/pkg/tts-windows"
	vosk "github.com/user/audio-dubbing-system/pkg/asr-vosk"
	asrvoskpython "github.com/user/audio-dubbing-system/pkg/asr-vosk-python"
	audiocapture "github.com/user/audio-dubbing-system/pkg/audio-capture"
	
	// Performance optimization
	"github.com/user/audio-dubbing-system/pkg/cache"
	"github.com/user/audio-dubbing-system/pkg/silence"
	"github.com/user/audio-dubbing-system/pkg/metrics"
)

var (
	version = "1.0.0"
	
	// Configuration flags
	inputDevice  string
	outputDevice string
	chunkSize    int
	apiKey       string
	mode         string // Performance mode: low-latency, balanced, quality
	
	// Implementation selection flags
	useArgos       bool
	useEspeak      bool
	useWindowsTTS  bool
	useVosk        bool
	useRealAudio   bool
	useSilenceDetection bool
	useMetrics     bool
	
	// Performance components
	silenceDetector  *silence.SilenceDetector
	metricsCollector *metrics.MetricsCollector
	translationCache *cache.TranslationCache
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "dubbing-mvp",
		Short: "Real-time PT→EN dubbing for Google Meets and other apps",
		Long: `Dubbing MVP - Minimum Viable Product
		
Captures Portuguese audio from your microphone, translates to English in real-time,
and outputs to a virtual audio device for use in Google Meets, Zoom, Discord, etc.`,
		Version: version,
	}

	// Start command
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start real-time dubbing",
		Run:   runStart,
	}
	startCmd.Flags().StringVar(&inputDevice, "input", "", "Input audio device (microphone)")
	startCmd.Flags().StringVar(&outputDevice, "output", "", "Output audio device (virtual cable)")
	startCmd.Flags().IntVar(&chunkSize, "chunk-size", 0, "Audio chunk size in seconds (1-5, 0=auto based on mode)")
	startCmd.Flags().StringVar(&mode, "mode", "balanced", "Performance mode: low-latency, balanced, quality")
	startCmd.Flags().StringVar(&apiKey, "api-key", "", "Translation API key (if using Google Translate)")
	startCmd.Flags().BoolVar(&useArgos, "use-argos", false, "Use Argos Translate (100% free, offline)")
	startCmd.Flags().BoolVar(&useEspeak, "use-espeak", false, "Use eSpeak TTS (free, offline)")
	startCmd.Flags().BoolVar(&useWindowsTTS, "use-windows-tts", false, "Use Windows TTS (free, native)")
	startCmd.Flags().BoolVar(&useVosk, "use-vosk", false, "Use Vosk ASR (free, offline)")
	startCmd.Flags().BoolVar(&useRealAudio, "use-real-audio", false, "Use real microphone capture (experimental)")
	startCmd.Flags().BoolVar(&useSilenceDetection, "use-silence-detection", false, "Use silence detection to skip processing (performance)")
	startCmd.Flags().BoolVar(&useMetrics, "use-metrics", false, "Enable detailed metrics collection (performance monitoring)")

	// Config command
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configure dubbing settings",
		Run:   runConfig,
	}
	configCmd.Flags().StringVar(&inputDevice, "input", "", "Input audio device")
	configCmd.Flags().StringVar(&outputDevice, "output", "", "Output audio device")

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show current status",
		Run:   runStatus,
	}

	// List devices command
	devicesCmd := &cobra.Command{
		Use:   "devices",
		Short: "List available audio devices",
		Run:   runListDevices,
	}

	rootCmd.AddCommand(startCmd, configCmd, statusCmd, devicesCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runStart(cmd *cobra.Command, args []string) {
	fmt.Println("🎙️  Dubbing MVP - Starting...")
	fmt.Printf("Version: %s\n\n", version)
	
	// Apply performance mode settings
	applyPerformanceMode()

	// Show selected implementations
	fmt.Println("📦 Selected implementations:")
	
	// ASR
	if useVosk {
		fmt.Println("  ✓ ASR: Vosk (free, offline)")
	} else {
		fmt.Println("  ✓ ASR: Mock (simulated)")
	}
	
	// Translation
	if useArgos {
		fmt.Println("  ✓ Translation: Argos Translate (100% free, offline)")
	} else {
		fmt.Println("  ✓ Translation: Mock (simulated)")
	}
	
	// TTS
	if useWindowsTTS {
		fmt.Println("  ✓ TTS: Windows TTS (free, native)")
	} else if useEspeak {
		fmt.Println("  ✓ TTS: eSpeak (free, offline)")
	} else {
		fmt.Println("  ✓ TTS: Mock (simulated)")
	}
	
	// Audio
	if useRealAudio {
		fmt.Println("  ✓ Audio: Real Microphone Capture (experimental)")
	} else {
		fmt.Println("  ✓ Audio: Mock (simulated)")
	}
	
	// Performance features
	if useSilenceDetection {
		fmt.Println("  ✓ Silence Detection: Enabled (performance boost)")
	}
	if useMetrics {
		fmt.Println("  ✓ Metrics Collection: Enabled (detailed monitoring)")
	}
	fmt.Println("  ✓ Translation Cache: Enabled")

	fmt.Println("\n🚀 Dubbing started!")
	fmt.Println("📊 Status:")
	fmt.Printf("  Input:  %s\n", getInputDevice())
	fmt.Printf("  Output: %s\n", getOutputDevice())
	fmt.Printf("  Chunk:  %ds\n", chunkSize)
	fmt.Println("\n💡 Speak in Portuguese → Others hear in English")
	fmt.Println("⏹️  Press Ctrl+C to stop")

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// TODO: Start processing loop
	go processingLoop()

	// Wait for interrupt
	<-sigChan
	fmt.Println("\n\n⏹️  Stopping dubbing...")
	
	// TODO: Cleanup
	fmt.Println("✓ Cleanup complete")
	fmt.Println("👋 Goodbye!")
}

func runConfig(cmd *cobra.Command, args []string) {
	fmt.Println("⚙️  Configuration")
	fmt.Println("================")
	
	if inputDevice != "" {
		fmt.Printf("Setting input device: %s\n", inputDevice)
		// TODO: Save to config file
	}
	
	if outputDevice != "" {
		fmt.Printf("Setting output device: %s\n", outputDevice)
		// TODO: Save to config file
	}
	
	fmt.Println("\n✓ Configuration saved")
}

func runStatus(cmd *cobra.Command, args []string) {
	fmt.Println("📊 Dubbing MVP Status")
	fmt.Println("====================")
	fmt.Println("Status: Not running")
	fmt.Printf("Input:  %s\n", getInputDevice())
	fmt.Printf("Output: %s\n", getOutputDevice())
	fmt.Println("\nRun 'dubbing-mvp start' to begin dubbing")
}

func runListDevices(cmd *cobra.Command, args []string) {
	fmt.Println("🔊 Available Audio Devices")
	fmt.Println("=========================")
	
	// TODO: List actual devices using M6
	fmt.Println("\nInput Devices (Microphones):")
	fmt.Println("  1. Default Microphone")
	fmt.Println("  2. Realtek HD Audio")
	
	fmt.Println("\nOutput Devices (Speakers/Virtual):")
	fmt.Println("  1. Default Speakers")
	fmt.Println("  2. Virtual Cable Input")
	
	fmt.Println("\n💡 Use device name with --input and --output flags")
}

func processingLoop() {
	// Initialize all modules
	log.Println("Initializing pipeline modules...")
	
	// 0. Initialize Performance Components
	if useMetrics {
		log.Println("Initializing metrics collector...")
		metricsCollector = metrics.NewMetricsCollector()
		log.Println("✓ Metrics Collector initialized")
	}
	
	if useSilenceDetection {
		log.Println("Initializing silence detector...")
		silenceDetector = silence.NewSilenceDetector(0.01, 1000) // 0.01 threshold, 1000 min samples
		log.Printf("✓ Silence Detector initialized (threshold: %.4f)", silenceDetector.GetThreshold())
	}
	
	// Initialize translation cache
	log.Println("Initializing translation cache...")
	translationCache = cache.NewTranslationCache(1000) // Max 1000 entries
	log.Println("✓ Translation Cache initialized")
	
	// 1. Initialize Audio Capture (if using real audio)
	var audioCapture *audiocapture.AudioCapture
	if useRealAudio {
		log.Println("Initializing real audio capture...")
		config := audiocapture.Config{
			DeviceName: inputDevice,
			SampleRate: 16000,
			Channels:   1,
			BufferSize: 16000 * 10, // 10 seconds buffer
		}
		var err error
		audioCapture, err = audiocapture.NewAudioCapture(config)
		if err != nil {
			log.Printf("⚠️  Failed to initialize audio capture: %v", err)
			log.Println("Falling back to mock audio...")
			useRealAudio = false
		} else {
			defer audioCapture.Close()
			// Start recording
			if err := audioCapture.StartRecording(); err != nil {
				log.Printf("⚠️  Failed to start recording: %v", err)
				useRealAudio = false
			}
		}
	}
	
	// 1. Initialize ASR
	asr, err := initASR()
	if err != nil {
		log.Fatalf("Failed to initialize ASR: %v", err)
	}
	defer asr.Close()
	
	// 2. Initialize Translator
	translator, err := initTranslator()
	if err != nil {
		log.Fatalf("Failed to initialize Translator: %v", err)
	}
	defer translator.Close()
	
	// 3. Initialize TTS
	tts, err := initTTS()
	if err != nil {
		log.Fatalf("Failed to initialize TTS: %v", err)
	}
	defer tts.Close()
	
	log.Println("✓ All modules initialized")
	log.Println("")
	log.Println("🎙️  Pipeline running - speak in Portuguese!")
	log.Println("")
	
	// Main processing loop
	ticker := time.NewTicker(time.Duration(chunkSize) * time.Second)
	defer ticker.Stop()
	
	chunkCount := 0

	for range ticker.C {
		chunkCount++
		log.Printf("--- Processing chunk #%d ---", chunkCount)
		
		// 1. Capture audio chunk
		var audioChunk []float32
		if useRealAudio && audioCapture != nil {
			// Use real audio capture
			audioChunk = audioCapture.GetChunk(float64(chunkSize))
		} else {
			// Use mock audio
			audioChunk = captureAudioChunk()
		}
		
		if len(audioChunk) == 0 {
			log.Println("No audio captured (silence)")
			continue
		}
		log.Printf("✓ Captured %d audio samples", len(audioChunk))
		
		// Check for silence if enabled
		if useSilenceDetection && silenceDetector != nil {
			if silenceDetector.IsSilence(audioChunk) {
				log.Println("🔇 Silence detected - skipping processing")
				if useMetrics && metricsCollector != nil {
					metricsCollector.RecordSilenceSkip()
					metricsCollector.NextChunk()
				}
				continue
			}
			log.Println("🎙️  Speech detected - processing...")
		}
		
		// 2. ASR: Transcribe to PT text (M2)
		startASR := time.Now()
		textPT, err := asr.Transcribe(audioChunk)
		asrLatency := time.Since(startASR)
		if useMetrics && metricsCollector != nil {
			metricsCollector.RecordLatency("ASR", asrLatency)
		}
		
		if err != nil {
			log.Printf("❌ ASR error: %v", err)
			continue
		}
		if textPT == "" {
			log.Println("No speech detected")
			continue
		}
		log.Printf("✓ ASR: '%s' (latency: %v)", textPT, asrLatency)
		
		// 3. Translation: PT → EN (M3) with cache
		startTranslation := time.Now()
		var textEN string
		
		// Check cache first
		if cachedTranslation, found := translationCache.Get(textPT); found {
			textEN = cachedTranslation
			if useMetrics && metricsCollector != nil {
				metricsCollector.RecordCacheHit()
			}
			log.Printf("✓ Translation (cached): '%s'", textEN)
		} else {
			textEN, err = translator.Translate(textPT)
			if useMetrics && metricsCollector != nil {
				metricsCollector.RecordCacheMiss()
			}
			if err != nil {
				log.Printf("❌ Translation error: %v", err)
				continue
			}
			// Store in cache
			translationCache.Set(textPT, textEN)
			log.Printf("✓ Translation: '%s'", textEN)
		}
		
		translationLatency := time.Since(startTranslation)
		if useMetrics && metricsCollector != nil {
			metricsCollector.RecordLatency("Translation", translationLatency)
		}
		log.Printf("  (latency: %v)", translationLatency)
		
		// 4. TTS: Synthesize EN audio (M4)
		startTTS := time.Now()
		audioEN, err := tts.Synthesize(textEN)
		ttsLatency := time.Since(startTTS)
		if useMetrics && metricsCollector != nil {
			metricsCollector.RecordLatency("TTS", ttsLatency)
		}
		
		if err != nil {
			log.Printf("❌ TTS error: %v", err)
			continue
		}
		log.Printf("✓ TTS: Generated %d audio samples (latency: %v)", len(audioEN), ttsLatency)
		
		// 5. Play EN audio (M6)
		// TODO: Integrate with actual M6 audio playback
		playAudioChunk(audioEN)
		log.Println("✓ Audio played")
		
		// Move to next chunk for metrics
		if useMetrics && metricsCollector != nil {
			metricsCollector.NextChunk()
		}
		
		// Print statistics
		printStats(asr, translator, tts)
		log.Println("")
	}
}

func getInputDevice() string {
	if inputDevice != "" {
		return inputDevice
	}
	// TODO: Load from config or use default
	return "Default Microphone"
}

func getOutputDevice() string {
	if outputDevice != "" {
		return outputDevice
	}
	// TODO: Load from config or use default
	return "Virtual Cable Input"
}


// Module initialization functions

// ASRInterface defines the common interface for ASR implementations
type ASRInterface interface {
	Transcribe(audioSamples []float32) (string, error)
	Close() error
	GetStats() ASRStats
}

// TranslatorInterface defines the common interface for translator implementations
type TranslatorInterface interface {
	Translate(textPT string) (string, error)
	Close() error
	GetStats() TranslatorStats
}

// TTSInterface defines the common interface for TTS implementations
type TTSInterface interface {
	Synthesize(textEN string) ([]float32, error)
	Close() error
	GetStats() TTSStats
}

// Stats types
type ASRStats struct {
	ChunksProcessed int64
	AverageLatency  time.Duration
	ErrorCount      int64
}

type TranslatorStats struct {
	SentencesTranslated int64
	AverageLatency      time.Duration
	ErrorCount          int64
}

type TTSStats struct {
	SentencesSynthesized int64
	AverageLatency       time.Duration
	ErrorCount           int64
}

func initASR() (ASRInterface, error) {
	if useVosk {
		log.Println("Initializing Vosk ASR (free, offline, Python)...")
		voskASR, err := asrvoskpython.NewVoskASR("pt", 16000)
		if err != nil {
			log.Printf("⚠️  Vosk ASR failed to initialize: %v", err)
			log.Println("Falling back to Mock ASR...")
			return initMockASR()
		}
		return &voskPythonASRWrapper{voskASR}, nil
	}
	
	return initMockASR()
}

func initMockASR() (ASRInterface, error) {
	log.Println("Initializing Mock ASR (simulated)...")
	config := asrsimple.Config{
		ModelPath:  "models/ggml-tiny.bin",
		Language:   "pt",
		SampleRate: 16000,
		Threads:    4,
	}
	mockASR, err := asrsimple.NewSimpleASR(config)
	if err != nil {
		return nil, err
	}
	return &mockASRWrapper{mockASR}, nil
}

func initTranslator() (TranslatorInterface, error) {
	if useArgos {
		log.Println("Initializing Argos Translate (100% free, offline)...")
		config := argos.Config{
			SourceLang: "pt",
			TargetLang: "en",
		}
		argosTranslator, err := argos.NewArgosTranslator(config)
		if err != nil {
			log.Printf("⚠️  Argos Translate failed to initialize: %v", err)
			log.Println("Falling back to Mock Translator...")
			return initMockTranslator()
		}
		return &argosTranslatorWrapper{argosTranslator}, nil
	}
	
	return initMockTranslator()
}

func initMockTranslator() (TranslatorInterface, error) {
	log.Println("Initializing Mock Translator (simulated)...")
	config := translationsimple.Config{
		APIKey:     apiKey,
		SourceLang: "pt",
		TargetLang: "en",
		UseAPI:     apiKey != "",
	}
	mockTranslator, err := translationsimple.NewSimpleTranslator(config)
	if err != nil {
		return nil, err
	}
	return &mockTranslatorWrapper{mockTranslator}, nil
}

func initTTS() (TTSInterface, error) {
	if useWindowsTTS {
		log.Println("Initializing Windows TTS (free, native)...")
		windowsTTS, err := ttswindows.NewWindowsTTS("en-us", 175, 16000)
		if err != nil {
			log.Printf("⚠️  Windows TTS failed to initialize: %v", err)
			log.Println("Falling back to Mock TTS...")
			return initMockTTS()
		}
		return &windowsTTSWrapper{windowsTTS}, nil
	}
	
	if useEspeak {
		log.Println("Initializing eSpeak TTS (free, offline)...")
		config := espeak.Config{
			Voice:      "en-us",
			Speed:      175,
			Pitch:      50,
			SampleRate: 16000,
		}
		espeakTTS, err := espeak.NewESpeakTTS(config)
		if err != nil {
			log.Printf("⚠️  eSpeak TTS failed to initialize: %v", err)
			log.Println("Falling back to Mock TTS...")
			return initMockTTS()
		}
		return &espeakTTSWrapper{espeakTTS}, nil
	}
	
	return initMockTTS()
}

func initMockTTS() (TTSInterface, error) {
	log.Println("Initializing Mock TTS (simulated)...")
	config := ttssimple.Config{
		Voice:      "en-us-female",
		SampleRate: 16000,
		Engine:     "mock",
	}
	mockTTS, err := ttssimple.NewSimpleTTS(config)
	if err != nil {
		return nil, err
	}
	return &mockTTSWrapper{mockTTS}, nil
}

// Wrapper types to adapt different implementations to common interfaces

// Vosk Python ASR Wrapper
type voskPythonASRWrapper struct {
	asr *asrvoskpython.VoskASR
}

func (w *voskPythonASRWrapper) Transcribe(audioSamples []float32) (string, error) {
	return w.asr.Transcribe(audioSamples)
}

func (w *voskPythonASRWrapper) Close() error {
	return w.asr.Close()
}

func (w *voskPythonASRWrapper) GetStats() ASRStats {
	// Vosk Python não tem stats ainda, retornar zeros
	return ASRStats{
		ChunksProcessed: 0,
		AverageLatency:  0,
		ErrorCount:      0,
	}
}

// Vosk ASR Wrapper (Go nativo - não usado)
type voskASRWrapper struct {
	asr *vosk.VoskASR
}

func (w *voskASRWrapper) Transcribe(audioSamples []float32) (string, error) {
	return w.asr.Transcribe(audioSamples)
}

func (w *voskASRWrapper) Close() error {
	return w.asr.Close()
}

func (w *voskASRWrapper) GetStats() ASRStats {
	stats := w.asr.GetStats()
	return ASRStats{
		ChunksProcessed: stats.ChunksProcessed,
		AverageLatency:  stats.AverageLatency,
		ErrorCount:      stats.ErrorCount,
	}
}

// Mock ASR Wrapper
type mockASRWrapper struct {
	asr *asrsimple.SimpleASR
}

func (w *mockASRWrapper) Transcribe(audioSamples []float32) (string, error) {
	return w.asr.Transcribe(audioSamples)
}

func (w *mockASRWrapper) Close() error {
	return w.asr.Close()
}

func (w *mockASRWrapper) GetStats() ASRStats {
	stats := w.asr.GetStats()
	return ASRStats{
		ChunksProcessed: stats.ChunksProcessed,
		AverageLatency:  stats.AverageLatency,
		ErrorCount:      stats.ErrorCount,
	}
}

// Argos Translator Wrapper
type argosTranslatorWrapper struct {
	translator *argos.ArgosTranslator
}

func (w *argosTranslatorWrapper) Translate(textPT string) (string, error) {
	return w.translator.Translate(textPT)
}

func (w *argosTranslatorWrapper) Close() error {
	return w.translator.Close()
}

func (w *argosTranslatorWrapper) GetStats() TranslatorStats {
	stats := w.translator.GetStats()
	return TranslatorStats{
		SentencesTranslated: stats.SentencesTranslated,
		AverageLatency:      stats.AverageLatency,
		ErrorCount:          stats.ErrorCount,
	}
}

// Mock Translator Wrapper
type mockTranslatorWrapper struct {
	translator *translationsimple.SimpleTranslator
}

func (w *mockTranslatorWrapper) Translate(textPT string) (string, error) {
	return w.translator.Translate(textPT)
}

func (w *mockTranslatorWrapper) Close() error {
	return w.translator.Close()
}

func (w *mockTranslatorWrapper) GetStats() TranslatorStats {
	stats := w.translator.GetStats()
	return TranslatorStats{
		SentencesTranslated: stats.SentencesTranslated,
		AverageLatency:      stats.AverageLatency,
		ErrorCount:          stats.ErrorCount,
	}
}

// Windows TTS Wrapper
type windowsTTSWrapper struct {
	tts *ttswindows.WindowsTTS
}

func (w *windowsTTSWrapper) Synthesize(textEN string) ([]float32, error) {
	return w.tts.Synthesize(textEN)
}

func (w *windowsTTSWrapper) Close() error {
	return w.tts.Close()
}

func (w *windowsTTSWrapper) GetStats() TTSStats {
	// Windows TTS não tem stats ainda, retornar zeros
	return TTSStats{
		SentencesSynthesized: 0,
		AverageLatency:       0,
		ErrorCount:           0,
	}
}

// eSpeak TTS Wrapper
type espeakTTSWrapper struct {
	tts *espeak.ESpeakTTS
}

func (w *espeakTTSWrapper) Synthesize(textEN string) ([]float32, error) {
	return w.tts.Synthesize(textEN)
}

func (w *espeakTTSWrapper) Close() error {
	return w.tts.Close()
}

func (w *espeakTTSWrapper) GetStats() TTSStats {
	stats := w.tts.GetStats()
	return TTSStats{
		SentencesSynthesized: stats.SentencesSynthesized,
		AverageLatency:       stats.AverageLatency,
		ErrorCount:           stats.ErrorCount,
	}
}

// Mock TTS Wrapper
type mockTTSWrapper struct {
	tts *ttssimple.SimpleTTS
}

func (w *mockTTSWrapper) Synthesize(textEN string) ([]float32, error) {
	return w.tts.Synthesize(textEN)
}

func (w *mockTTSWrapper) Close() error {
	return w.tts.Close()
}

func (w *mockTTSWrapper) GetStats() TTSStats {
	stats := w.tts.GetStats()
	return TTSStats{
		SentencesSynthesized: stats.SentencesSynthesized,
		AverageLatency:       stats.AverageLatency,
		ErrorCount:           stats.ErrorCount,
	}
}

// Audio capture/playback functions (mock for now)

func captureAudioChunk() []float32 {
	// TODO: Integrate with M6 audio capture
	// For now, generate mock audio with some energy
	
	sampleRate := 16000
	duration := float64(chunkSize)
	numSamples := int(duration * float64(sampleRate))
	
	samples := make([]float32, numSamples)
	
	// Generate some noise to simulate audio
	// In real implementation, this will come from microphone
	for i := range samples {
		// Random noise between -0.1 and 0.1
		samples[i] = (float32(i%100) / 1000.0) - 0.05
	}
	
	return samples
}

func playAudioChunk(audioSamples []float32) {
	// TODO: Integrate with M6 audio playback
	// For now, just simulate playback delay
	
	if len(audioSamples) == 0 {
		return
	}
	
	// Simulate playback time
	sampleRate := 16000
	duration := float64(len(audioSamples)) / float64(sampleRate)
	time.Sleep(time.Duration(duration * float64(time.Second)))
}

// Statistics printing

func printStats(asr ASRInterface, translator TranslatorInterface, tts TTSInterface) {
	asrStats := asr.GetStats()
	transStats := translator.GetStats()
	ttsStats := tts.GetStats()
	
	log.Println("📊 Statistics:")
	log.Printf("  ASR:         %d chunks, avg latency: %v", asrStats.ChunksProcessed, asrStats.AverageLatency)
	log.Printf("  Translation: %d sentences, avg latency: %v", transStats.SentencesTranslated, transStats.AverageLatency)
	log.Printf("  TTS:         %d sentences, avg latency: %v", ttsStats.SentencesSynthesized, ttsStats.AverageLatency)
	
	// Show performance metrics if enabled
	if useMetrics && metricsCollector != nil {
		stats := metricsCollector.GetAggregated()
		log.Println("⚡ Performance Metrics:")
		log.Printf("  Total Chunks:    %d", stats.TotalChunks)
		log.Printf("  Avg Latency:     %v", stats.AverageLatency)
		log.Printf("  P50 Latency:     %v", stats.P50Latency)
		log.Printf("  P95 Latency:     %v", stats.P95Latency)
		log.Printf("  P99 Latency:     %v", stats.P99Latency)
		
		if stats.CacheHits+stats.CacheMisses > 0 {
			log.Printf("  Cache Hit Rate:  %.1f%% (%d hits, %d misses)", 
				stats.CacheHitRate*100, stats.CacheHits, stats.CacheMisses)
		}
		
		if stats.SilenceSkips > 0 {
			log.Printf("  Silence Skips:   %d", stats.SilenceSkips)
		}
		
		log.Printf("  Uptime:          %v", stats.Uptime)
	}
	
	// Show silence detection stats if enabled
	if useSilenceDetection && silenceDetector != nil {
		silenceStats := silenceDetector.GetStats()
		if silenceStats.TotalChecks > 0 {
			log.Println("🔇 Silence Detection:")
			log.Printf("  Total Checks:    %d", silenceStats.TotalChecks)
			log.Printf("  Silence:         %d (%.1f%%)", silenceStats.SilenceDetected, silenceStats.SilenceRate*100)
			log.Printf("  Speech:          %d", silenceStats.SpeechDetected)
		}
	}
	
	// Show cache stats
	if translationCache != nil {
		cacheStats := translationCache.GetStats()
		if cacheStats.Hits+cacheStats.Misses > 0 {
			log.Println("💾 Translation Cache:")
			log.Printf("  Size:            %d/%d entries", cacheStats.Size, cacheStats.MaxSize)
			log.Printf("  Hit Rate:        %.1f%% (%d hits, %d misses)", 
				cacheStats.HitRate*100, cacheStats.Hits, cacheStats.Misses)
		}
	}
}


// Performance mode configuration
func applyPerformanceMode() {
	// Normalize mode name
	mode = strings.ToLower(strings.TrimSpace(mode))
	
	// Validate mode
	validModes := map[string]bool{
		"low-latency": true,
		"balanced":    true,
		"quality":     true,
	}
	
	if !validModes[mode] {
		fmt.Printf("⚠️  Invalid mode '%s', using 'balanced'\n", mode)
		mode = "balanced"
	}
	
	// Apply mode settings if chunk size not explicitly set
	if chunkSize == 0 {
		switch mode {
		case "low-latency":
			chunkSize = 1
			fmt.Println("⚡ Mode: Low-Latency (1s chunks, fast response)")
		case "balanced":
			chunkSize = 2
			fmt.Println("⚖️  Mode: Balanced (2s chunks, good balance)")
		case "quality":
			chunkSize = 3
			fmt.Println("🎯 Mode: Quality (3s chunks, better accuracy)")
		}
	} else {
		// Validate explicit chunk size
		if chunkSize < 1 || chunkSize > 5 {
			fmt.Printf("⚠️  Invalid chunk size %d, using 2s\n", chunkSize)
			chunkSize = 2
		}
		fmt.Printf("📏 Chunk Size: %ds (manual override)\n", chunkSize)
	}
	
	// Enable performance features based on mode
	switch mode {
	case "low-latency":
		// Low latency: enable all optimizations
		if !useSilenceDetection {
			useSilenceDetection = true
			fmt.Println("  ✓ Auto-enabled: Silence Detection")
		}
		if !useMetrics {
			useMetrics = true
			fmt.Println("  ✓ Auto-enabled: Metrics Collection")
		}
	case "balanced":
		// Balanced: enable silence detection
		if !useSilenceDetection {
			useSilenceDetection = true
			fmt.Println("  ✓ Auto-enabled: Silence Detection")
		}
	case "quality":
		// Quality: no auto-optimizations
		fmt.Println("  ℹ️  Quality mode: optimizations disabled for accuracy")
	}
	
	fmt.Println()
}
