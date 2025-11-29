# Design Document: Automatic Dubbing System

## Overview

The Automatic Dubbing System is a real-time streaming pipeline that transforms Portuguese speech into English speech while preserving the speaker's voice characteristics, prosody, and semantic meaning. The system operates with a strict latency budget of 700ms end-to-end to enable natural conversational flow.

The architecture follows a modular pipeline design with six primary modules:
1. **Audio Capture & Preprocessing** - Captures and cleans Portuguese audio input
2. **ASR (Automatic Speech Recognition)** - Converts Portuguese audio to text with timestamps
3. **Translation** - Translates Portuguese text to English while preserving semantics
4. **TTS (Text-to-Speech)** - Synthesizes English speech with the user's voice
5. **Synchronization & Streaming** - Manages temporal alignment and continuous flow
6. **Audio Output** - Delivers final audio to speakers/headphones

Each module operates incrementally on streaming data, emitting results as soon as they're available rather than waiting for complete utterances.

## Architecture

### High-Level Pipeline Flow

```
[Microphone] → [M1: Capture] → [M2: ASR] → [M3: Translation] → [M4: TTS] → [M5: Sync] → [M6: Output] → [Speakers]
```

### Latency Budget Allocation

- **M1 (Capture & Preprocessing)**: 50ms
- **M2 (ASR)**: 200ms
- **M3 (Translation)**: 150ms
- **M4 (TTS)**: 200ms
- **M5 (Synchronization)**: 50ms
- **M6 (Audio Output)**: 30ms
- **Buffer overhead**: 20ms
- **Total**: 700ms

### Data Flow Architecture

The system uses a **push-based streaming architecture** where each module:
- Receives data chunks from the previous module
- Processes incrementally without waiting for complete utterances
- Emits results immediately to the next module
- Implements backpressure when downstream modules are overloaded

### Error Handling Strategy

- **Graceful degradation**: Modules continue processing even when individual chunks fail
- **Logging and monitoring**: All errors are logged with context for debugging
- **Automatic recovery**: Modules attempt to recover from transient failures
- **Quality flags**: Low-confidence outputs are flagged but not blocked

## Components and Interfaces

### M1: Audio Capture & Preprocessing

**Responsibilities:**
- Capture raw audio from microphone
- Apply noise reduction and normalization
- Detect voice activity (VAD)
- Segment audio into PCM frames (10-20ms)
- Manage ring buffer to prevent overflow

**Interface:**
```go
type AudioCaptureModule interface {
    Start() error
    Stop() error
    GetFrameStream() <-chan PCMFrame
    SetBackpressure(enabled bool)
}

type PCMFrame struct {
    Data        []int16      // Audio samples
    SampleRate  int          // Samples per second (e.g., 16000)
    Timestamp   time.Time    // Capture timestamp
    Duration    time.Duration // Frame duration
    IsSpeech    bool         // VAD result
}
```

**Key Components:**
- **Microphone Interface**: Platform-specific audio capture (ALSA, CoreAudio, WASAPI)
- **Noise Reduction**: Spectral subtraction or Wiener filtering
- **VAD**: Energy-based or ML-based voice activity detection
- **Ring Buffer**: Circular buffer with configurable size (default 500ms capacity)

### M2: ASR (Automatic Speech Recognition)

**Responsibilities:**
- Convert Portuguese audio to text incrementally
- Generate token-level timestamps
- Emit partial hypotheses for streaming
- Handle disfluencies and corrections

**Interface:**
```go
type ASRModule interface {
    ProcessFrame(frame PCMFrame) error
    GetTokenStream() <-chan ASRToken
    Reset() error
}

type ASRToken struct {
    Text       string
    Language   string        // "pt-BR"
    Timestamp  time.Time     // Aligned with audio
    Confidence float64       // 0.0 - 1.0
    IsFinal    bool          // True if hypothesis is finalized
}
```

**Implementation Options:**
- **Whisper** (OpenAI): High accuracy, supports streaming with modifications
- **Vosk**: Lightweight, fully offline, good for real-time
- **Google Cloud Speech-to-Text**: Streaming API with excellent accuracy
- **Azure Speech Services**: Low latency streaming mode

**Recommended**: Vosk for offline/low-latency or Whisper with streaming adapter

### M3: Translation Module

**Responsibilities:**
- Translate Portuguese tokens to English
- Preserve semantic meaning and style
- Maintain context across sentences
- Add prosody markers for TTS

**Interface:**
```go
type TranslationModule interface {
    Translate(tokens []ASRToken) ([]TranslatedToken, error)
    GetContextWindow() int
    SetContextWindow(size int)
}

type TranslatedToken struct {
    SourceText    string
    TargetText    string
    SourceLang    string        // "pt-BR"
    TargetLang    string        // "en-US"
    Timestamp     time.Time
    Confidence    float64
    ProsodyMarker ProsodyInfo   // Duration, emphasis hints
}

type ProsodyInfo struct {
    RelativeDuration float64  // 0.8 = faster, 1.2 = slower
    EmphasisLevel    int      // 0 = none, 1 = moderate, 2 = strong
    PauseAfter       time.Duration
}
```

**Sub-components:**

#### M3.1: Literal Translation
- Uses neural machine translation (NMT) model
- Options: MarianMT, NLLB (Meta), Google Translate API, DeepL API
- Recommended: NLLB-200 for offline or DeepL API for quality

#### M3.2: Semantic Equivalence
- Validates translation preserves meaning using sentence embeddings
- Computes cosine similarity between source and target embeddings
- Uses models like: Sentence-BERT, LaBSE (Language-agnostic BERT)
- Threshold: 0.85 minimum similarity

#### M3.3: Prosody Adjustment
- Analyzes source prosody from ASR timestamps
- Adjusts target text phrasing for natural TTS output
- Inserts pause markers and emphasis hints

### M4: TTS (Text-to-Speech) with Voice Cloning

**Responsibilities:**
- Generate English speech with user's voice characteristics
- Maintain prosody and emotional tone
- Stream audio chunks with low latency

**Interface:**
```go
type TTSModule interface {
    SetSpeakerEmbedding(embedding []float32) error
    Synthesize(tokens []TranslatedToken) error
    GetAudioStream() <-chan AudioChunk
}

type AudioChunk struct {
    Data       []int16
    SampleRate int
    Timestamp  time.Time
    Duration   time.Duration
}
```

**Sub-components:**

#### M4.1: Voice Cloning
- Extracts speaker embedding from user's Portuguese samples
- Options:
  - **Coqui TTS** (XTTS model): Open-source, good quality
  - **Tortoise TTS**: High quality but slower
  - **ElevenLabs API**: Commercial, excellent quality
  - **Azure Neural TTS** (Custom Neural Voice): Commercial
- Recommended: Coqui XTTS for offline or ElevenLabs for quality

#### M4.2: Prosody Control
- Adjusts pitch envelope based on source prosody
- Controls speaking rate and rhythm
- Applies emphasis and emotional coloring

#### M4.3: Streaming Generation
- Generates audio in small chunks (50-100ms)
- Uses streaming vocoder (e.g., HiFi-GAN, WaveGlow)
- Maintains continuous output without gaps

### M5: Synchronization & Streaming

**Responsibilities:**
- Align Portuguese source timestamps with English output
- Manage buffering to prevent underruns
- Maintain continuous audio flow
- Monitor end-to-end latency

**Interface:**
```go
type SyncModule interface {
    AddChunk(chunk AudioChunk) error
    GetSynchronizedStream() <-chan AudioChunk
    GetLatencyMetrics() LatencyMetrics
}

type LatencyMetrics struct {
    EndToEndLatency time.Duration
    BufferFillLevel float64  // 0.0 - 1.0
    DroppedChunks   int
}
```

**Sub-components:**

#### M5.1: Buffering
- Circular buffer with 300ms capacity
- Jitter absorption
- Overflow/underflow protection

#### M5.2: Temporal Alignment
- Maps source timestamps to target timestamps
- Accounts for translation expansion/compression
- Maintains 50ms alignment accuracy

#### M5.3: Latency Monitoring
- Tracks end-to-end latency
- Logs violations of 700ms budget
- Triggers optimization when needed

### M6: Audio Output

**Responsibilities:**
- Interface with OS audio drivers
- Mix multiple audio channels if needed
- Ensure continuous playback without glitches

**Interface:**
```go
type AudioOutputModule interface {
    Start() error
    Stop() error
    WriteAudio(chunk AudioChunk) error
    GetOutputLatency() time.Duration
}
```

**Sub-components:**

#### M6.1: Audio Drivers

**Platform-specific implementations:**
- **Linux**: ALSA, PulseAudio, JACK
- **macOS**: CoreAudio
- **Windows**: WASAPI (focus of initial implementation)

**Target latency**: ≤30ms for capture, ≤50ms for playback

**Windows Implementation (M6.1-Win) - Detailed Design:**

##### M6.1.1-Win: Audio Capture
**Responsibility**: Receive audio from user's microphone in continuous PCM frames

**Components**:
- API: WASAPI (Windows Audio Session API) in Exclusive or Shared mode
- Circular capture buffer (10-20ms per frame)
- Event-driven callback for frame delivery
- Thread isolation via goroutine

**Interface**:
```go
type WindowsAudioCapture interface {
    Initialize(deviceID string, sampleRate int, channels int) error
    Start() error
    Stop() error
    GetFrameChannel() <-chan PCMFrame
    GetCaptureLatency() time.Duration
}
```

**Validation Criteria**:
- Each received frame has expected size (10-20ms)
- Continuity test: no frames lost during 10s capture
- End-to-end latency ≤ 30ms (internal buffer + driver)

##### M6.1.2-Win: Audio Playback
**Responsibility**: Play synthesized English audio smoothly without gaps

**Components**:
- API: WASAPI, Exclusive mode preferred for low latency
- Adaptive circular output buffer (40-80ms)
- Jitter buffer to compensate for timing variations
- Thread isolation via goroutine

**Interface**:
```go
type WindowsAudioPlayback interface {
    Initialize(deviceID string, sampleRate int, channels int) error
    Start() error
    Stop() error
    WriteFrame(frame PCMFrame) error
    GetPlaybackLatency() time.Duration
    GetBufferFillLevel() float64
}
```

**Validation Criteria**:
- Continuous synchronization without underruns
- End-to-end latency (frame to output) ≤ 50ms
- Buffer flush and continuous resume test

##### M6.1.3-Win: Stream Synchronization
**Responsibility**: Maintain temporal alignment between capture and playback

**Components**:
- Per-frame timestamps
- Clock drift compensation
- Adaptive buffer feedback

**Interface**:
```go
type StreamSynchronizer interface {
    SyncCapturePlayback(captureTime, playbackTime time.Time) error
    GetDriftCompensation() time.Duration
    AdjustBufferSize(targetLatency time.Duration) error
}
```

**Validation Criteria**:
- Drift test: maximum 1-2ms per second
- Load simulation: buffers don't overflow with parallel processing

##### M6.1.4-Win: Latency Management
**Responsibility**: Optimize buffers and operation modes to achieve <150ms end-to-end latency target

**Components**:
- Dynamic buffer size adjustment based on CPU load
- Exclusive vs Shared mode selection
- Adaptive jitter compensation

**Interface**:
```go
type LatencyManager interface {
    MonitorLatency() LatencyMetrics
    OptimizeBuffers(cpuLoad float64) error
    SelectOperationMode() (WASAPIMode, error)
}

type WASAPIMode int
const (
    Exclusive WASAPIMode = iota
    Shared
)
```

**Validation Criteria**:
- End-to-end latency benchmark under different conditions (idle, high CPU load)
- Jitter logs ≤ 10ms

**Operation Flow**:
```
[Microphone] → (M6.1.1 Capture) → ring buffer → processing (ASR/MT/TTS)
                                        ↓
                            (M6.1.3 Synchronization)
                                        ↓
                            (M6.1.2 Playback) → [Speakers]
```

**Threading Model**:
- Each sub-module operates in separate goroutine
- Communication via lock-free channels or queues with backpressure
- WASAPI callback hooks for capture and playback frames

**Implementation Libraries**:
- Primary: `github.com/moutend/go-wasapi` (native WASAPI Go bindings)
- Alternative: `github.com/gordonklaus/portaudio` (cross-platform PortAudio)
- Ring buffers: Custom implementation or `github.com/smallnest/ringbuffer`

#### M6.2: Mixing Pipeline
- Real-time sample rate conversion if needed
- Volume normalization
- Multi-channel mixing

#### M6.3: Continuous Output
- Prevents buffer underruns
- Inserts silence on starvation rather than replaying
- Monitors for glitches

## Data Models

### Core Data Structures

```go
// PCMFrame represents a chunk of audio data
type PCMFrame struct {
    Data        []int16       // 16-bit PCM samples
    SampleRate  int           // Typically 16000 Hz
    Channels    int           // Mono = 1, Stereo = 2
    Timestamp   time.Time     // Capture time
    Duration    time.Duration // Frame duration (10-20ms)
    IsSpeech    bool          // VAD result
}

// ASRToken represents a recognized word/token
type ASRToken struct {
    Text       string
    Language   string
    Timestamp  time.Time
    Duration   time.Duration
    Confidence float64
    IsFinal    bool
}

// TranslatedToken represents translated text with metadata
type TranslatedToken struct {
    SourceText    string
    TargetText    string
    SourceLang    string
    TargetLang    string
    Timestamp     time.Time
    Confidence    float64
    SemanticScore float64      // Cosine similarity
    ProsodyMarker ProsodyInfo
}

// AudioChunk represents synthesized audio
type AudioChunk struct {
    Data       []int16
    SampleRate int
    Channels   int
    Timestamp  time.Time
    Duration   time.Duration
    SourceRef  string  // Reference to source token
}

// PipelineMetrics tracks system performance
type PipelineMetrics struct {
    CaptureLatency      time.Duration
    ASRLatency          time.Duration
    TranslationLatency  time.Duration
    TTSLatency          time.Duration
    SyncLatency         time.Duration
    OutputLatency       time.Duration
    EndToEndLatency     time.Duration
    DroppedFrames       int
    ErrorCount          int
}
```

### State Management

Each module maintains minimal internal state:
- **Capture**: Ring buffer state, VAD state
- **ASR**: Partial hypothesis, acoustic model state
- **Translation**: Context window (last 3 sentences)
- **TTS**: Speaker embedding, vocoder state
- **Sync**: Buffer contents, timestamp mappings
- **Output**: Playback position, driver state

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property 1: Frame Duration Consistency
*For any* captured audio input, all emitted PCM frames should have duration between 10-20ms and be in valid PCM format.
**Validates: Requirements 1.1**

### Property 2: Noise Reduction Effectiveness
*For any* audio input containing background noise, the noise reduction processing should maintain signal-to-noise ratio above -20dB.
**Validates: Requirements 1.2**

### Property 3: VAD Response Time
*For any* speech input containing pauses, the VAD should detect silence transitions within 50ms and correctly segment the audio stream.
**Validates: Requirements 1.3**

### Property 4: Pipeline Continuity
*For any* streaming audio session, the system should maintain continuous frame/chunk delivery without buffer overflows, underflows, or audible gaps throughout the entire pipeline.
**Validates: Requirements 1.4, 5.2, 6.4**

### Property 5: ASR Streaming Latency
*For any* PCM frame input to the ASR engine, partial text hypotheses should be emitted within 200ms.
**Validates: Requirements 2.1**

### Property 6: Temporal Alignment Accuracy
*For any* recognized token or synthesized audio chunk, timestamps should align with the source audio within the specified accuracy threshold (10ms for ASR tokens, 50ms for final output).
**Validates: Requirements 2.2, 5.3**

### Property 7: Incremental Token Emission
*For any* speech recognition operation, tokens should be emitted immediately upon recognition without waiting for sentence completion.
**Validates: Requirements 2.3**

### Property 8: Translation Latency Bound
*For any* set of Portuguese tokens from the ASR engine, corresponding English tokens should be generated within 150ms.
**Validates: Requirements 3.1**

### Property 9: Semantic Preservation
*For any* translated segment, the cosine similarity between source Portuguese and target English semantic embeddings should be at least 0.85.
**Validates: Requirements 3.2, 3.3**

### Property 10: Context Window Maintenance
*For any* translation operation, the system should maintain a context window of at least 3 previous sentences for coherent translation.
**Validates: Requirements 3.4**

### Property 11: Prosody Annotation Completeness
*For any* translated token output, the token should include prosody markers (duration, emphasis, pause information).
**Validates: Requirements 3.5**

### Property 12: Speaker Embedding Application
*For any* TTS synthesis operation, the system should use the speaker embedding derived from the user's voice samples.
**Validates: Requirements 4.1**

### Property 13: Prosody Preservation
*For any* speech synthesis with emotional inflection in the source, the output should preserve pitch variations within 10% of the source prosody envelope and match timing within 50ms.
**Validates: Requirements 4.3, 4.5**

### Property 14: Buffer Capacity Maintenance
*For any* audio streaming session, the circular buffer should maintain capacity for at least 300ms of audio.
**Validates: Requirements 5.1**

### Property 15: End-to-End Latency Bound
*For any* complete pipeline execution from Portuguese audio input to English audio output, the total latency should not exceed 700ms.
**Validates: Requirements 5.4**

### Property 16: Audio Output Latency
*For any* audio chunk sent to the output device, the driver latency should not exceed 30ms.
**Validates: Requirements 6.1**

### Property 17: Mixing Quality
*For any* multi-channel audio mixing operation, the output should not contain phase issues or clipping artifacts.
**Validates: Requirements 6.2**

### Property 18: Resampling Quality
*For any* sample rate conversion operation, the resampled audio should not introduce aliasing or other perceptible artifacts.
**Validates: Requirements 6.3**

### Property 19: Error Logging and Recovery
*For any* error encountered by any module, the system should log the error with timestamp and context, and continue processing subsequent data rather than blocking the pipeline.
**Validates: Requirements 7.1, 7.2**

### Property 20: Metrics Exposure
*For any* module in the system, performance metrics (latency, throughput, error rates) should be exposed through a standardized monitoring interface.
**Validates: Requirements 8.5**

## Error Handling

### Error Categories

1. **Transient Errors**: Temporary issues that may resolve (network timeouts, brief audio glitches)
   - Strategy: Retry with exponential backoff, continue processing

2. **Quality Degradation**: Output quality below threshold (low ASR confidence, poor translation similarity)
   - Strategy: Flag the segment, log warning, continue processing

3. **Resource Exhaustion**: Buffer overflow, memory pressure, CPU saturation
   - Strategy: Apply backpressure, drop non-critical data, alert monitoring

4. **Fatal Errors**: Module crashes, unrecoverable failures
   - Strategy: Attempt automatic restart, failover to backup if available

### Error Handling Principles

- **Never block the pipeline**: Errors in one chunk should not prevent processing of subsequent chunks
- **Graceful degradation**: Reduce quality rather than fail completely
- **Comprehensive logging**: All errors logged with full context for debugging
- **Automatic recovery**: Modules attempt self-recovery before escalating
- **User notification**: Critical errors that affect output quality are surfaced to users

### Error Propagation

Errors are handled locally within modules and do not propagate upstream. Each module:
- Catches and logs errors
- Attempts recovery or workaround
- Emits best-effort output or skips the problematic chunk
- Updates error metrics for monitoring

## Testing Strategy

### Dual Testing Approach

The system will employ both **unit testing** and **property-based testing** to ensure comprehensive coverage:

- **Unit tests** verify specific examples, edge cases, and integration points
- **Property tests** verify universal properties hold across all inputs
- Together they provide complete validation: unit tests catch concrete bugs, property tests verify general correctness

### Unit Testing

Unit tests will cover:

1. **Module Integration Points**
   - Data flow between modules
   - Interface contract compliance
   - Error propagation

2. **Specific Edge Cases**
   - Empty audio input
   - Very long utterances
   - Rapid speaker changes
   - Silence-only input
   - Maximum buffer capacity

3. **Error Conditions**
   - Module initialization failures
   - Invalid input formats
   - Resource exhaustion scenarios
   - Recovery from crashes

### Property-Based Testing

**Framework**: We will use **gopter** (Go property testing library) for implementing property-based tests.

**Configuration**: Each property test will run a minimum of **100 iterations** to ensure statistical confidence in the results.

**Test Tagging**: Each property-based test will include a comment explicitly referencing the correctness property from this design document using the format:
```go
// Feature: auto-dubbing-system, Property 1: Frame Duration Consistency
```

**Property Test Coverage**:

Each of the 20 correctness properties defined above will be implemented as a property-based test. The tests will:

1. Generate random valid inputs for the module under test
2. Execute the module operation
3. Verify the property holds for the output
4. Report any counterexamples that violate the property

**Example Property Test Structure**:

```go
// Feature: auto-dubbing-system, Property 1: Frame Duration Consistency
func TestProperty_FrameDurationConsistency(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("all frames have duration 10-20ms", prop.ForAll(
        func(audioInput []int16) bool {
            frames := captureModule.Process(audioInput)
            for _, frame := range frames {
                duration := frame.Duration.Milliseconds()
                if duration < 10 || duration > 20 {
                    return false
                }
            }
            return true
        },
        gen.SliceOf(gen.Int16()),
    ))
    
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

### Integration Testing

Integration tests will validate:
- End-to-end pipeline flow with real audio samples
- Latency measurements under various load conditions
- Quality metrics (semantic similarity, prosody preservation)
- Error recovery scenarios

### Performance Testing

Performance tests will measure:
- Per-module latency under normal and peak load
- End-to-end latency distribution
- Throughput (audio minutes processed per minute)
- Resource utilization (CPU, memory, GPU if applicable)

### Quality Validation

Quality tests will assess:
- Semantic similarity scores on test translation pairs
- Prosody preservation metrics
- Voice similarity (using automated voice comparison tools)
- Subjective quality through limited human evaluation

## Implementation Notes

### Technology Stack Recommendations

**Language**: Go (for performance, concurrency, and cross-platform support)

**Key Libraries**:
- **Audio I/O**: PortAudio (cross-platform) or platform-specific APIs
- **ASR**: Vosk (offline, low latency) or Whisper with streaming modifications
- **Translation**: NLLB-200 (offline) or DeepL API (cloud)
- **TTS**: Coqui TTS (XTTS model) for voice cloning
- **Embeddings**: Sentence-Transformers for semantic similarity
- **Testing**: gopter for property-based testing, standard Go testing for unit tests

### Performance Optimization Strategies

1. **Parallel Processing**: Run independent modules in parallel where possible
2. **Batch Processing**: Process multiple frames together when latency allows
3. **Model Quantization**: Use quantized models (INT8) for faster inference
4. **GPU Acceleration**: Offload ML models to GPU when available
5. **Caching**: Cache speaker embeddings and frequently used translations
6. **Streaming**: Use streaming APIs for all ML models to reduce latency

### Deployment Considerations

- **Resource Requirements**: Estimate 4-8 GB RAM, modern CPU (or GPU for optimal performance)
- **Offline Capability**: Design for offline operation with local models
- **Scalability**: Support multiple concurrent sessions if needed
- **Monitoring**: Integrate with observability tools (Prometheus, Grafana)
- **Configuration**: Externalize model paths, latency budgets, quality thresholds

### Future Enhancements

- Support for additional language pairs
- Adaptive quality based on available resources
- Multi-speaker scenarios
- Real-time quality feedback to users
- Cloud-based model fallback for improved quality
