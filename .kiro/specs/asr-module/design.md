# Design Document: ASR Module

## Overview

The ASR (Automatic Speech Recognition) Module is responsible for converting streaming Portuguese audio into timestamped Portuguese text tokens. The module receives PCM audio frames from the Audio Interface Module (M6), processes them through a multi-stage pipeline, and emits ASR tokens to the Translation Module (M3).

### Key Design Goals

1. **Low Latency**: Maintain end-to-end latency below 200ms per chunk
2. **High Accuracy**: Achieve Word Error Rate (WER) < 15% for Brazilian Portuguese
3. **Real-Time Performance**: Process audio faster than real-time (RTF < 0.5)
4. **Streaming Support**: Emit partial and final results incrementally
5. **Robustness**: Handle errors gracefully without pipeline disruption
6. **Observability**: Provide comprehensive metrics for monitoring

### Architecture Principles

Following the SOLID principles demonstrated in the Audio Interface Module:

- **Single Responsibility**: Each sub-module has one clear responsibility
- **Open/Closed**: Extensible through interfaces without modifying existing code
- **Liskov Substitution**: Interfaces allow swapping implementations
- **Interface Segregation**: Focused interfaces for specific contracts
- **Dependency Inversion**: Depend on abstractions, not concrete implementations

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      ASR Module (M2)                         │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐      ┌──────────────┐      ┌───────────┐ │
│  │   Feature    │─────►│  Recognition │─────►│  Decoder  │ │
│  │  Extractor   │      │    Engine    │      │           │ │
│  └──────────────┘      └──────────────┘      └───────────┘ │
│         │                      │                     │       │
│         ▼                      ▼                     ▼       │
│  ┌──────────────┐      ┌──────────────┐      ┌───────────┐ │
│  │    Audio     │      │    Model     │      │  Language │ │
│  │  Normalizer  │      │    Loader    │      │   Model   │ │
│  └──────────────┘      └──────────────┘      └───────────┘ │
│         │                                            │       │
│         ▼                                            ▼       │
│  ┌──────────────┐                          ┌────────────┐  │
│  │     VAD      │                          │    Text    │  │
│  │   Enhancer   │                          │ Normalizer │  │
│  └──────────────┘                          └────────────┘  │
│                                                    │         │
│  ┌──────────────────────────────────────────┐    │         │
│  │         Chunk Manager                     │    │         │
│  │  ┌────────────┐      ┌────────────┐     │    │         │
│  │  │  Context   │      │  Partial   │     │    │         │
│  │  │   Window   │      │ Hypothesis │     │    │         │
│  │  └────────────┘      └────────────┘     │    │         │
│  └──────────────────────────────────────────┘    │         │
│                                                    │         │
│  ┌──────────────────────────────────────────┐    │         │
│  │         Post-Processing                   │◄───┘         │
│  │  ┌────────────┐      ┌────────────┐     │              │
│  │  │ Timestamp  │      │ Confidence │     │              │
│  │  │  Aligner   │      │   Scorer   │     │              │
│  │  └────────────┘      └────────────┘     │              │
│  └──────────────────────────────────────────┘              │
│                          │                                   │
│                          ▼                                   │
│  ┌──────────────────────────────────────────┐              │
│  │         ASR Coordinator                   │              │
│  │  - Orchestrates all sub-modules           │              │
│  │  - Manages lifecycle                      │              │
│  │  - Handles errors                         │              │
│  │  - Collects metrics                       │              │
│  └──────────────────────────────────────────┘              │
│                          │                                   │
└──────────────────────────┼───────────────────────────────────┘
                           │
                           ▼
                   Token Stream Output
```

### Module Decomposition

```
M2: ASR Module
├── M2.1: Audio Preprocessing
│   ├── M2.1.1: FeatureExtractor (MFCC/Mel-spectrogram extraction)
│   ├── M2.1.2: AudioNormalizer (Amplitude normalization)
│   └── M2.1.3: VADEnhancer (Voice activity detection)
│
├── M2.2: Recognition Engine
│   ├── M2.2.1: ModelLoader (Load and manage ASR model)
│   ├── M2.2.2: InferenceEngine (Execute model inference)
│   └── M2.2.3: BeamSearchDecoder (Decode logits to text)
│
├── M2.3: Streaming Management
│   ├── M2.3.1: ChunkManager (Divide audio into chunks)
│   ├── M2.3.2: ContextWindow (Maintain context between chunks)
│   └── M2.3.3: PartialHypothesis (Emit incremental results)
│
├── M2.4: Post-Processing
│   ├── M2.4.1: TextNormalizer (Normalize punctuation, capitalization)
│   ├── M2.4.2: TimestampAligner (Align words with audio)
│   └── M2.4.3: ConfidenceScorer (Compute confidence scores)
│
├── M2.5: Language Model Integration
│   ├── M2.5.1: LanguageModelLoader (Load language model)
│   └── M2.5.2: LMRescorer (Rescore hypotheses with LM)
│
├── M2.6: Observability
│   ├── M2.6.1: LatencyTracker (Track processing latency)
│   └── M2.6.2: AccuracyMonitor (Estimate WER in real-time)
│
└── M2.7: Orchestration
    └── M2.7.1: ASRCoordinator (Coordinate all sub-modules)
```

## Components and Interfaces

### Core Interfaces

```go
// ASRModule is the main interface for the ASR system
type ASRModule interface {
    // Lifecycle management
    Initialize(config ASRConfig) error
    Start() error
    Stop() error
    Close() error
    
    // Core functionality
    ProcessFrame(frame types.PCMFrame) error
    GetTokenStream() <-chan ASRToken
    
    // Configuration
    SetLanguage(lang string) error
    SetModel(modelPath string) error
    
    // Observability
    GetLatency() time.Duration
    GetAccuracy() float64
    GetStats() ASRStats
}

// AudioPreprocessor handles audio preprocessing
type AudioPreprocessor interface {
    ExtractFeatures(frame types.PCMFrame) (Features, error)
    Normalize(frame types.PCMFrame) types.PCMFrame
    DetectVoiceActivity(frame types.PCMFrame) bool
}

// RecognitionEngine performs speech recognition
type RecognitionEngine interface {
    Initialize(modelPath string) error
    Infer(features Features) (Logits, error)
    Decode(logits Logits, beamSize int) ([]Hypothesis, error)
    GetRTF() float64
}

// StreamingManager handles chunking and streaming
type StreamingManager interface {
    AddFrame(frame types.PCMFrame) error
    GetNextChunk() (Chunk, bool)
    MaintainContext(chunk Chunk)
    EmitPartialHypothesis(text string, confidence float64)
}

// PostProcessor handles text post-processing
type PostProcessor interface {
    NormalizeText(text string) string
    AlignTimestamps(text string, audio []types.PCMFrame) []WordTiming
    ComputeConfidence(hypothesis Hypothesis) float64
}

// LanguageModelIntegration handles LM rescoring
type LanguageModelIntegration interface {
    Initialize(lmPath string) error
    Rescore(hypotheses []Hypothesis) []Hypothesis
}

// MetricsCollector collects ASR metrics
type MetricsCollector interface {
    RecordLatency(duration time.Duration)
    RecordRTF(rtf float64)
    EstimateWER(confidence float64)
    GetStats() ASRStats
}
```

### Sub-Module Interfaces

```go
// FeatureExtractor extracts acoustic features
type FeatureExtractor interface {
    ExtractMFCC(frame types.PCMFrame) ([]float32, error)
    ExtractMelSpectrogram(frame types.PCMFrame) ([][]float32, error)
    GetFeatureDimension() int
}

// AudioNormalizer normalizes audio amplitude
type AudioNormalizer interface {
    Normalize(frame types.PCMFrame) types.PCMFrame
    RemoveDCOffset(frame types.PCMFrame) types.PCMFrame
    GetNormalizationFactor() float32
}

// VADEnhancer detects voice activity
type VADEnhancer interface {
    DetectVoiceActivity(frame types.PCMFrame) bool
    GetEnergyThreshold() float32
    SetEnergyThreshold(threshold float32)
}

// ModelLoader loads and manages ASR models
type ModelLoader interface {
    LoadModel(path string) error
    UnloadModel() error
    GetModelInfo() ModelInfo
    IsModelLoaded() bool
}

// InferenceEngine executes model inference
type InferenceEngine interface {
    Infer(features Features) (Logits, error)
    SetBatchSize(size int)
    GetInferenceTime() time.Duration
}

// BeamSearchDecoder decodes logits to text
type BeamSearchDecoder interface {
    Decode(logits Logits, beamSize int) ([]Hypothesis, error)
    SetBeamSize(size int)
    GetBeamSize() int
}

// ChunkManager manages audio chunks
type ChunkManager interface {
    AddFrame(frame types.PCMFrame) error
    GetNextChunk() (Chunk, bool)
    SetChunkSize(duration time.Duration)
    GetChunkSize() time.Duration
}

// ContextWindow maintains context
type ContextWindow interface {
    AddChunk(chunk Chunk)
    GetContext() []Chunk
    SetWindowSize(size int)
    Clear()
}

// PartialHypothesis emits partial results
type PartialHypothesis interface {
    EmitPartial(text string, confidence float64, isFinal bool)
    GetPartialStream() <-chan ASRToken
}

// TextNormalizer normalizes text
type TextNormalizer interface {
    NormalizePunctuation(text string) string
    NormalizeCapitalization(text string) string
    NormalizeWhitespace(text string) string
}

// TimestampAligner aligns timestamps
type TimestampAligner interface {
    AlignWords(text string, audio []types.PCMFrame) []WordTiming
    GetAlignmentPrecision() time.Duration
}

// ConfidenceScorer computes confidence
type ConfidenceScorer interface {
    ComputeWordConfidence(hypothesis Hypothesis) []float64
    ComputeUtteranceConfidence(wordConfidences []float64) float64
}

// LanguageModelLoader loads language models
type LanguageModelLoader interface {
    LoadLanguageModel(path string) error
    UnloadLanguageModel() error
    IsLanguageModelLoaded() bool
}

// LMRescorer rescores with language model
type LMRescorer interface {
    Rescore(hypotheses []Hypothesis) []Hypothesis
    SetLMWeight(weight float64)
    GetLMWeight() float64
}

// LatencyTracker tracks latency
type LatencyTracker interface {
    RecordLatency(duration time.Duration)
    GetAverageLatency() time.Duration
    GetP95Latency() time.Duration
}

// AccuracyMonitor estimates accuracy
type AccuracyMonitor interface {
    EstimateWER(confidence float64)
    GetEstimatedWER() float64
    GetAverageConfidence() float64
}
```

## Data Models

### Core Data Types

```go
// ASRToken represents a recognized text token
type ASRToken struct {
    Text         string        // Recognized text
    Language     string        // Language code (e.g., "pt-BR")
    Timestamp    time.Time     // Start time of the token
    Duration     time.Duration // Duration of the audio segment
    Confidence   float64       // Confidence score (0.0 to 1.0)
    IsFinal      bool          // Whether this is a final result
    WordTimings  []WordTiming  // Word-level timing information
}

// WordTiming represents timing for a single word
type WordTiming struct {
    Word       string        // The word text
    Start      time.Time     // Start time of the word
    End        time.Time     // End time of the word
    Confidence float64       // Confidence score for this word
}

// ASRConfig holds configuration for the ASR module
type ASRConfig struct {
    ModelPath      string        // Path to ASR model
    Language       string        // Language code (e.g., "pt-BR")
    SampleRate     int           // Audio sample rate (e.g., 16000)
    ChunkSize      time.Duration // Size of audio chunks (e.g., 2s)
    ContextWindow  int           // Number of previous chunks to maintain
    BeamSize       int           // Beam size for decoding (e.g., 5)
    LanguageModel  string        // Path to language model (optional)
    EnableVAD      bool          // Enable voice activity detection
    EnableLM       bool          // Enable language model rescoring
}

// Features represents extracted audio features
type Features struct {
    Type       FeatureType   // MFCC or MelSpectrogram
    Data       [][]float32   // Feature matrix [time][feature_dim]
    SampleRate int           // Original sample rate
    FrameCount int           // Number of frames
}

// FeatureType specifies the type of features
type FeatureType int

const (
    MFCC FeatureType = iota
    MelSpectrogram
)

// Logits represents model output
type Logits struct {
    Data       [][]float32   // Logit matrix [time][vocab_size]
    VocabSize  int           // Size of vocabulary
    FrameCount int           // Number of time frames
}

// Hypothesis represents a recognition hypothesis
type Hypothesis struct {
    Text            string    // Recognized text
    AcousticScore   float64   // Acoustic model score
    LanguageScore   float64   // Language model score
    CombinedScore   float64   // Combined score
    WordConfidences []float64 // Per-word confidence scores
    Tokens          []int     // Token IDs
}

// Chunk represents an audio chunk
type Chunk struct {
    Frames    []types.PCMFrame // Audio frames in this chunk
    StartTime time.Time        // Start time of the chunk
    Duration  time.Duration    // Duration of the chunk
    HasVoice  bool             // Whether chunk contains voice
}

// ModelInfo contains model metadata
type ModelInfo struct {
    Name         string
    Version      string
    Language     string
    SampleRate   int
    VocabSize    int
    Architecture string
}

// ASRStats contains performance statistics
type ASRStats struct {
    // Latency metrics
    ProcessingLatency time.Duration // Average processing latency
    P95Latency        time.Duration // 95th percentile latency
    RTF               float64       // Real-Time Factor
    
    // Accuracy metrics
    EstimatedWER      float64       // Estimated Word Error Rate
    ConfidenceAvg     float64       // Average confidence score
    
    // Throughput metrics
    FramesProcessed   int64         // Total frames processed
    TokensEmitted     int64         // Total tokens emitted
    ChunksProcessed   int64         // Total chunks processed
    
    // Error metrics
    ErrorCount        int64         // Total errors
    DroppedFrames     int64         // Frames dropped due to errors
    InferenceFailures int64         // Model inference failures
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property 1: Feature Extraction Produces Valid Dimensions

*For any* PCM frame with valid audio data, extracting MFCC or Mel-Spectrogram features should produce a feature matrix with dimensions matching the configured feature type and frame count.

**Validates: Requirements 1.1**

### Property 2: Audio Normalization Maintains Range

*For any* audio frame regardless of input amplitude, normalizing the frame should produce output samples within the standard normalized range [-1.0, 1.0].

**Validates: Requirements 1.2**

### Property 3: VAD Correctly Classifies Voice Activity

*For any* audio frame, applying voice activity detection should return a boolean result, and frames with energy above the threshold should be classified as containing voice.

**Validates: Requirements 1.3**

### Property 4: Model Inference Produces Valid Logits

*For any* valid feature input, executing model inference should produce logits with dimensions [time_frames, vocab_size] where vocab_size matches the model vocabulary.

**Validates: Requirements 2.2**

### Property 5: Beam Search Produces Multiple Hypotheses

*For any* logit input and beam size N, beam search decoding should produce at most N hypotheses, each with a valid combined score.

**Validates: Requirements 2.3**

### Property 6: Real-Time Factor Stays Below Threshold

*For any* audio chunk processed, the Real-Time Factor (processing_time / audio_duration) should be less than 0.5 to maintain real-time performance.

**Validates: Requirements 2.5**

### Property 7: Chunking Produces Correct Sizes

*For any* continuous audio stream, dividing into chunks should produce chunks with durations within the configured range [1s, 3s], except possibly the final chunk.

**Validates: Requirements 3.1**

### Property 8: Context Window Maintains Fixed Size

*For any* sequence of chunks processed, the context window should maintain exactly the configured number of previous chunks (e.g., 2 chunks).

**Validates: Requirements 3.2**

### Property 9: Partial Hypotheses Precede Final Results

*For any* chunk processed, if partial hypotheses are emitted, they should be emitted before the final result for that chunk.

**Validates: Requirements 3.3**

### Property 10: Partial Hypothesis Flag Correctness

*For any* partial hypothesis emitted, the IsFinal flag should be false, and for any final result emitted, the IsFinal flag should be true.

**Validates: Requirements 3.4, 3.5**

### Property 11: Text Normalization Preserves Content

*For any* recognized text, normalizing punctuation and capitalization should preserve the semantic content while improving formatting.

**Validates: Requirements 4.1, 4.2**

### Property 12: Timestamp Alignment Precision

*For any* word in recognized text, the aligned timestamp should be within 50 milliseconds of the actual audio position.

**Validates: Requirements 4.3**

### Property 13: Confidence Scores Within Valid Range

*For any* word or utterance, the computed confidence score should be within the range [0.0, 1.0].

**Validates: Requirements 4.4**

### Property 14: Aggregate Confidence Derived from Words

*For any* utterance with word-level confidence scores, the aggregate utterance confidence should be computed as a function (e.g., average) of the word confidences.

**Validates: Requirements 4.5**

### Property 15: Language Model Rescoring Changes Scores

*For any* set of hypotheses, applying language model rescoring should update the combined scores, potentially reordering the hypotheses.

**Validates: Requirements 5.2**

### Property 16: Best Hypothesis Selection

*For any* set of rescored hypotheses, the selected hypothesis should have the highest combined acoustic and language model score.

**Validates: Requirements 5.3**

### Property 17: Token Completeness

*For any* ASR token emitted, the token should contain all required fields: text, language, timestamp, duration, confidence, IsFinal flag, and word timings.

**Validates: Requirements 6.2**

### Property 18: Word Timings Match Text

*For any* ASR token emitted, the number of word timings should equal the number of words in the text, and each word timing should correspond to a word in the text.

**Validates: Requirements 6.3**

### Property 19: Backpressure Prevents Overflow

*For any* situation where the token stream channel is full, the ASR module should pause audio processing until channel capacity is available, preventing frame loss.

**Validates: Requirements 6.5**

### Property 20: Latency Tracking Completeness

*For any* chunk processed, the processing latency should be recorded in the metrics, and the average latency should be computable from all recorded values.

**Validates: Requirements 7.1**

### Property 21: RTF Calculation Correctness

*For any* processing operation, the Real-Time Factor should be calculated as processing_time / audio_duration, and should be a positive number.

**Validates: Requirements 7.2**

### Property 22: Counter Monotonicity

*For any* sequence of operations, the counters for frames processed, tokens emitted, and errors should be monotonically increasing (never decrease).

**Validates: Requirements 7.4, 7.5**

### Property 23: Error Recovery Continuity

*For any* corrupted frame or inference failure, the ASR module should log the error, handle it gracefully, and continue processing subsequent frames without crashing.

**Validates: Requirements 8.2, 8.3**

### Property 24: Frame Reception Latency

*For any* PCM frame emitted by the Audio Interface Module, the ASR module should receive and begin processing the frame within 10 milliseconds.

**Validates: Requirements 10.2**

### Property 25: Backpressure Prevents Buffer Overflow

*For any* situation where the ASR module processes frames slower than they are produced, backpressure should be applied to prevent buffer overflow and frame loss.

**Validates: Requirements 10.3**

## Error Handling

### Error Categories

```go
type ASRError struct {
    Code      ErrorCode
    Message   string
    Module    string
    Timestamp time.Time
    Context   map[string]interface{}
    Cause     error
}

type ErrorCode int

const (
    // Initialization errors
    ErrModelLoadFailed ErrorCode = iota
    ErrLanguageModelLoadFailed
    ErrInvalidConfig
    ErrResourceAllocation
    
    // Processing errors
    ErrFeatureExtraction
    ErrInferenceFailed
    ErrDecodingFailed
    ErrCorruptedFrame
    
    // Resource errors
    ErrMemoryExhaustion
    ErrChannelFull
    ErrTimeout
    
    // Integration errors
    ErrAudioInterfaceDisconnected
    ErrTranslationModuleUnavailable
)
```

### Error Handling Strategies

1. **Initialization Errors**: Fail fast and return error to caller
   - Model loading failures → return error, do not start
   - Invalid configuration → return error with details
   - Resource allocation failures → cleanup and return error

2. **Processing Errors**: Log and continue
   - Corrupted frames → log error, skip frame, continue
   - Feature extraction failures → log error, emit empty token, continue
   - Inference failures → log error, emit empty token, continue

3. **Resource Errors**: Attempt recovery
   - Memory exhaustion → clear buffers, trigger GC, attempt to continue
   - Channel full → apply backpressure, wait for capacity
   - Timeout → log warning, continue with partial results

4. **Degraded State**: After repeated failures
   - 3 consecutive inference failures → enter degraded state
   - Notify coordinator of degraded state
   - Continue attempting processing with reduced expectations
   - Attempt automatic recovery after cooldown period

### Error Recovery Flow

```
┌─────────────────┐
│  Normal State   │
└────────┬────────┘
         │
         ▼
    ┌────────┐
    │ Error  │
    └───┬────┘
        │
        ├─► Initialization Error ──► Fail Fast ──► Return Error
        │
        ├─► Processing Error ──► Log & Continue ──► Normal State
        │
        ├─► Resource Error ──► Attempt Recovery ──┬─► Success ──► Normal State
        │                                          └─► Failure ──► Degraded State
        │
        └─► Repeated Failures ──► Degraded State ──┬─► Cooldown ──► Retry
                                                    └─► Notify Coordinator
```

## Testing Strategy

### Unit Testing

Unit tests verify individual components in isolation:

1. **FeatureExtractor Tests**
   - Test MFCC extraction with known audio samples
   - Test Mel-spectrogram extraction with known audio samples
   - Test feature dimensions match expected values
   - Test edge cases: empty frames, very short frames, very long frames

2. **AudioNormalizer Tests**
   - Test normalization with various amplitude levels
   - Test DC offset removal
   - Test preservation of audio characteristics
   - Test edge cases: silent audio, clipped audio

3. **VADEnhancer Tests**
   - Test voice activity detection with speech samples
   - Test voice activity detection with silence
   - Test threshold adjustment
   - Test edge cases: very quiet speech, background noise

4. **ModelLoader Tests**
   - Test successful model loading
   - Test model loading failure handling
   - Test model metadata extraction
   - Test model unloading and cleanup

5. **InferenceEngine Tests**
   - Test inference with valid features
   - Test inference timing and RTF calculation
   - Test batch processing
   - Test edge cases: empty features, very long sequences

6. **BeamSearchDecoder Tests**
   - Test decoding with various beam sizes
   - Test hypothesis ranking
   - Test edge cases: empty logits, single-token sequences

7. **ChunkManager Tests**
   - Test chunk creation with various sizes
   - Test chunk boundary handling
   - Test edge cases: very short audio, very long audio

8. **ContextWindow Tests**
   - Test context maintenance
   - Test window size limits
   - Test context clearing
   - Test edge cases: empty context, single chunk

9. **TextNormalizer Tests**
   - Test punctuation normalization
   - Test capitalization rules
   - Test whitespace normalization
   - Test edge cases: empty text, special characters

10. **TimestampAligner Tests**
    - Test word-level alignment
    - Test alignment precision
    - Test edge cases: single word, very long utterances

11. **ConfidenceScorer Tests**
    - Test word confidence calculation
    - Test utterance confidence aggregation
    - Test edge cases: zero confidence, perfect confidence

12. **LanguageModelLoader Tests**
    - Test LM loading
    - Test LM loading failure handling
    - Test LM unloading

13. **LMRescorer Tests**
    - Test hypothesis rescoring
    - Test score combination
    - Test LM weight adjustment

14. **LatencyTracker Tests**
    - Test latency recording
    - Test average calculation
    - Test percentile calculation

15. **AccuracyMonitor Tests**
    - Test WER estimation
    - Test confidence tracking

### Property-Based Testing

Property-based tests verify universal properties across many random inputs:

1. **Property Test: Feature Extraction Dimensions**
   - Generate random PCM frames
   - Extract features
   - Verify feature dimensions are correct
   - **Validates: Property 1**

2. **Property Test: Normalization Range**
   - Generate random audio frames with various amplitudes
   - Normalize frames
   - Verify all samples are in [-1.0, 1.0]
   - **Validates: Property 2**

3. **Property Test: VAD Classification**
   - Generate random frames with various energy levels
   - Apply VAD
   - Verify classification is consistent with energy threshold
   - **Validates: Property 3**

4. **Property Test: Inference Output Dimensions**
   - Generate random feature inputs
   - Execute inference
   - Verify logit dimensions match [time, vocab_size]
   - **Validates: Property 4**

5. **Property Test: Beam Search Hypothesis Count**
   - Generate random logits
   - Apply beam search with various beam sizes
   - Verify hypothesis count ≤ beam size
   - **Validates: Property 5**

6. **Property Test: Chunking Size Correctness**
   - Generate random audio streams
   - Apply chunking
   - Verify chunk durations are within configured range
   - **Validates: Property 7**

7. **Property Test: Context Window Size**
   - Process random chunk sequences
   - Verify context window maintains exactly N previous chunks
   - **Validates: Property 8**

8. **Property Test: Partial Before Final**
   - Process random chunks
   - Verify partial hypotheses are emitted before final results
   - **Validates: Property 9**

9. **Property Test: IsFinal Flag Correctness**
   - Generate random hypotheses
   - Verify partial hypotheses have IsFinal=false
   - Verify final results have IsFinal=true
   - **Validates: Property 10**

10. **Property Test: Confidence Range**
    - Generate random recognition results
    - Verify all confidence scores are in [0.0, 1.0]
    - **Validates: Property 13**

11. **Property Test: Token Completeness**
    - Generate random tokens
    - Verify all required fields are present and valid
    - **Validates: Property 17**

12. **Property Test: Word Timings Match Text**
    - Generate random tokens with text
    - Verify word timing count equals word count
    - **Validates: Property 18**

13. **Property Test: Counter Monotonicity**
    - Process random sequences of operations
    - Verify counters never decrease
    - **Validates: Property 22**

14. **Property Test: Error Recovery Continuity**
    - Inject random errors during processing
    - Verify processing continues after errors
    - **Validates: Property 23**

### Integration Testing

Integration tests verify interactions between components:

1. **Audio Interface Integration**
   - Test receiving frames from Audio Interface Module
   - Test backpressure coordination
   - Test event handling (start/stop)
   - Verify end-to-end latency

2. **Translation Module Integration**
   - Test emitting tokens to Translation Module
   - Test channel communication
   - Test backpressure handling
   - Verify token format compatibility

3. **End-to-End Pipeline**
   - Test complete flow: audio → features → inference → decoding → tokens
   - Verify latency budget (< 200ms)
   - Verify accuracy on test dataset
   - Test with various audio conditions

4. **Concurrent Processing**
   - Test multiple concurrent audio streams
   - Verify thread safety
   - Test resource contention
   - Verify no race conditions

### Performance Testing

Performance tests verify latency and throughput requirements:

1. **Latency Tests**
   - Measure per-frame processing latency (target: < 50ms)
   - Measure per-chunk processing latency (target: < 200ms)
   - Measure end-to-end latency from frame receipt to token emission
   - Verify P95 latency meets requirements

2. **Throughput Tests**
   - Measure frames processed per second
   - Measure tokens emitted per second
   - Verify RTF < 0.5 (processes 2x faster than real-time)
   - Test sustained throughput over long durations

3. **Resource Usage Tests**
   - Measure CPU usage (target: < 30%)
   - Measure memory usage (target: < 500MB)
   - Measure GPU usage if applicable
   - Test memory leak detection over extended runs

4. **Stress Tests**
   - Test with maximum frame rate
   - Test with long-running sessions (hours)
   - Test with degraded resources
   - Test recovery from resource exhaustion

### Test Coverage Goals

- **Unit Test Coverage**: > 80% line coverage
- **Property Test Coverage**: All 25 correctness properties tested
- **Integration Test Coverage**: All module interfaces tested
- **Performance Test Coverage**: All latency and throughput requirements verified

## Implementation Notes

### Technology Stack

**ASR Model Options**:
1. **Whisper** (OpenAI)
   - Pros: High accuracy, multilingual, open source
   - Cons: Higher latency, larger model size
   - Recommended for: High accuracy requirements

2. **Vosk** (Alpha Cephei)
   - Pros: Low latency, small model size, offline
   - Cons: Lower accuracy than Whisper
   - Recommended for: Low latency requirements

3. **Wav2Vec 2.0** (Facebook)
   - Pros: Good accuracy, efficient
   - Cons: Requires fine-tuning for Portuguese
   - Recommended for: Custom training scenarios

**Language Model Options**:
1. **KenLM** - N-gram language model
2. **Transformer LM** - Neural language model
3. **GPT-based** - Large language model

**Feature Extraction**:
- Use `librosa` or `torchaudio` for MFCC/Mel-spectrogram extraction
- Consider GPU acceleration for feature extraction

**Inference Framework**:
- **ONNX Runtime** - Cross-platform, optimized
- **PyTorch** - Flexible, good for development
- **TensorFlow Lite** - Optimized for edge devices

### Performance Optimization

1. **Model Optimization**
   - Use quantized models (INT8) for faster inference
   - Use model pruning to reduce size
   - Consider distillation for smaller models

2. **Batching**
   - Batch multiple chunks for parallel inference
   - Use dynamic batching based on load

3. **Caching**
   - Cache model weights in memory
   - Cache language model states
   - Cache feature extraction parameters

4. **Parallelization**
   - Use worker pools for parallel chunk processing
   - Parallelize feature extraction
   - Use GPU for inference if available

5. **Memory Management**
   - Reuse buffers to reduce allocations
   - Use memory pools for frequent allocations
   - Implement aggressive garbage collection tuning

### Streaming Considerations

1. **Chunk Overlap**
   - Use 10-20% overlap between chunks to avoid boundary errors
   - Merge overlapping regions in post-processing

2. **Context Carryover**
   - Maintain hidden states between chunks for RNN-based models
   - Use attention masks for Transformer-based models

3. **Partial Results**
   - Emit partial results every 500ms for responsiveness
   - Mark partial results as non-final
   - Allow partial results to be updated

4. **Latency vs Accuracy Tradeoff**
   - Smaller chunks = lower latency, potentially lower accuracy
   - Larger chunks = higher latency, potentially higher accuracy
   - Tune chunk size based on requirements

### Integration Patterns

1. **Channel-Based Communication**
   ```go
   // Producer (ASR Module)
   tokenChannel := make(chan ASRToken, 100)
   go func() {
       for token := range recognizedTokens {
           tokenChannel <- token
       }
   }()
   
   // Consumer (Translation Module)
   for token := range tokenChannel {
       translate(token)
   }
   ```

2. **Backpressure Handling**
   ```go
   select {
   case tokenChannel <- token:
       // Token sent successfully
   case <-time.After(100 * time.Millisecond):
       // Channel full, apply backpressure
       pauseProcessing()
   }
   ```

3. **Error Propagation**
   ```go
   type Result struct {
       Token ASRToken
       Error error
   }
   
   resultChannel := make(chan Result, 100)
   ```

### Monitoring and Observability

1. **Metrics to Track**
   - Processing latency (avg, p50, p95, p99)
   - Real-Time Factor (RTF)
   - Estimated Word Error Rate
   - Throughput (frames/sec, tokens/sec)
   - Error rates by type
   - Resource usage (CPU, memory, GPU)

2. **Logging Strategy**
   - Log level: INFO for normal operations
   - Log level: WARN for recoverable errors
   - Log level: ERROR for serious failures
   - Include context: timestamp, module, operation
   - Use structured logging (JSON format)

3. **Tracing**
   - Trace request flow through pipeline
   - Measure time spent in each component
   - Identify bottlenecks
   - Use distributed tracing for multi-module systems

4. **Alerting**
   - Alert on high latency (> 200ms)
   - Alert on high error rate (> 5%)
   - Alert on degraded state
   - Alert on resource exhaustion

## Deployment Considerations

### Resource Requirements

**Minimum Requirements**:
- CPU: 4 cores
- RAM: 2GB
- Disk: 1GB for models
- Network: Not required (offline capable)

**Recommended Requirements**:
- CPU: 8 cores or GPU
- RAM: 4GB
- Disk: 5GB for models and cache
- Network: Optional for model updates

### Scalability

1. **Horizontal Scaling**
   - Run multiple ASR instances for parallel streams
   - Use load balancer to distribute streams
   - Share model cache across instances

2. **Vertical Scaling**
   - Use larger models for better accuracy
   - Use GPU for faster inference
   - Increase worker pool size

### Configuration Management

```yaml
asr:
  model:
    path: "/models/whisper-medium-pt.onnx"
    type: "whisper"
    language: "pt-BR"
  
  processing:
    sample_rate: 16000
    chunk_size: "2s"
    context_window: 2
    beam_size: 5
    enable_vad: true
  
  language_model:
    enabled: true
    path: "/models/pt-lm.bin"
    weight: 0.3
  
  performance:
    max_latency: "200ms"
    target_rtf: 0.5
    worker_pool_size: 4
  
  integration:
    audio_interface:
      frame_buffer_size: 100
      backpressure_threshold: 0.8
    translation:
      token_buffer_size: 100
```

## Future Enhancements

1. **Multi-Speaker Support**
   - Speaker diarization
   - Per-speaker models
   - Speaker-specific confidence scores

2. **Adaptive Chunk Sizing**
   - Dynamically adjust chunk size based on speech rate
   - Smaller chunks for fast speech
   - Larger chunks for slow speech

3. **Online Learning**
   - Adapt to user's voice over time
   - Learn user-specific vocabulary
   - Improve accuracy with usage

4. **Multi-Language Support**
   - Support multiple source languages
   - Automatic language detection
   - Language-specific optimizations

5. **Advanced Error Correction**
   - Use context from Translation Module for error correction
   - Implement confidence-based re-recognition
   - Use user feedback for model improvement
