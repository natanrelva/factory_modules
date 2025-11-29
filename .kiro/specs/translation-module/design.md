# Design Document: Translation Module

## Overview

The Translation Module (M3) is responsible for translating Portuguese text tokens into English text tokens while preserving semantic meaning, context, and prosodic information. The module sits between the ASR Module (M2) and TTS Module (M4) in the dubbing pipeline, receiving Portuguese text and emitting English text with quality assurance and prosodic annotations.

### Key Design Goals

1. **High Quality**: Achieve BLEU > 30 and semantic similarity > 0.85
2. **Low Latency**: Maintain translation latency < 150ms per sentence
3. **Context Awareness**: Use previous sentences for coherent translations
4. **Consistency**: Maintain terminology consistency across translations
5. **Prosody Transfer**: Preserve speech characteristics from source to target
6. **Robustness**: Handle errors gracefully without pipeline disruption

### Architecture Principles

Following SOLID principles from the Audio Interface Module:

- **Single Responsibility**: Each sub-module has one clear purpose
- **Open/Closed**: Extensible through interfaces
- **Liskov Substitution**: Swappable implementations via interfaces
- **Interface Segregation**: Focused, minimal interfaces
- **Dependency Inversion**: Depend on abstractions

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Translation Module (M3)                     │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐      ┌──────────────┐      ┌───────────┐ │
│  │     Text     │─────►│ Translation  │─────►│   Beam    │ │
│  │   Cleaner    │      │   Inference  │      │  Decoder  │ │
│  └──────────────┘      └──────────────┘      └───────────┘ │
│         │                      │                     │       │
│         ▼                      ▼                     ▼       │
│  ┌──────────────┐      ┌──────────────┐      ┌───────────┐ │
│  │  Tokenizer   │      │    Model     │      │  Quality  │ │
│  │              │      │    Loader    │      │ Validator │ │
│  └──────────────┘      └──────────────┘      └───────────┘ │
│         │                                            │       │
│         ▼                                            ▼       │
│  ┌──────────────┐                          ┌────────────┐  │
│  │   Sentence   │                          │  Semantic  │  │
│  │  Segmenter   │                          │ Validator  │  │
│  └──────────────┘                          └────────────┘  │
│                                                    │         │
│  ┌──────────────────────────────────────────┐    │         │
│  │         Context Management                │    │         │
│  │  ┌────────────┐      ┌────────────┐     │    │         │
│  │  │  Context   │      │Terminology │     │    │         │
│  │  │   Window   │      │   Cache    │     │    │         │
│  │  └────────────┘      └────────────┘     │    │         │
│  └──────────────────────────────────────────┘    │         │
│                                                    │         │
│  ┌──────────────────────────────────────────┐    │         │
│  │         Post-Processing                   │◄───┘         │
│  │  ┌────────────┐      ┌────────────┐     │              │
│  │  │Detokenizer │      │  Prosody   │     │              │
│  │  │            │      │ Annotator  │     │              │
│  │  └────────────┘      └────────────┘     │              │
│  └──────────────────────────────────────────┘              │
│                          │                                   │
│                          ▼                                   │
│  ┌──────────────────────────────────────────┐              │
│  │      Translation Coordinator              │              │
│  │  - Orchestrates all sub-modules           │              │
│  │  - Manages lifecycle                      │              │
│  │  - Handles errors                         │              │
│  │  - Collects metrics                       │              │
│  └──────────────────────────────────────────┘              │
│                          │                                   │
└──────────────────────────┼───────────────────────────────────┘
                           │
                           ▼
                 Translated Token Stream
```

### Module Decomposition

```
M3: Translation Module
├── M3.1: Text Preprocessing
│   ├── M3.1.1: TextCleaner (Remove noise, normalize)
│   ├── M3.1.2: Tokenizer (Subword tokenization)
│   └── M3.1.3: SentenceSegmenter (Sentence boundary detection)
│
├── M3.2: Translation Engine
│   ├── M3.2.1: ModelLoader (Load NMT model)
│   ├── M3.2.2: TranslationInference (Execute translation)
│   └── M3.2.3: BeamSearchDecoder (Generate hypotheses)
│
├── M3.3: Context Management
│   ├── M3.3.1: ContextWindow (Maintain sentence history)
│   ├── M3.3.2: TerminologyCache (Consistent term translation)
│   └── M3.3.3: ConversationState (Track topic/context)
│
├── M3.4: Quality Assurance
│   ├── M3.4.1: SemanticValidator (Verify semantic similarity)
│   ├── M3.4.2: FluencyScorer (Measure naturalness)
│   └── M3.4.3: LengthNormalizer (Adjust translation length)
│
├── M3.5: Post-Processing
│   ├── M3.5.1: Detokenizer (Reconstruct text)
│   ├── M3.5.2: ProsodyAnnotator (Add prosody markers)
│   └── M3.5.3: FormattingAdjuster (Fix punctuation/caps)
│
├── M3.6: Observability
│   ├── M3.6.1: LatencyTracker (Track translation latency)
│   └── M3.6.2: QualityMonitor (Track BLEU/similarity)
│
└── M3.7: Orchestration
    └── M3.7.1: TranslationCoordinator (Coordinate pipeline)
```

## Components and Interfaces

### Core Interfaces

```go
// TranslationModule is the main interface
type TranslationModule interface {
    // Lifecycle
    Initialize(config TranslationConfig) error
    Start() error
    Stop() error
    Close() error
    
    // Core functionality
    Translate(tokens []ASRToken) ([]TranslatedToken, error)
    TranslateStreaming(tokenStream <-chan ASRToken) <-chan TranslatedToken
    
    // Configuration
    SetSourceLanguage(lang string) error
    SetTargetLanguage(lang string) error
    SetContextWindow(size int) error
    
    // Observability
    GetLatency() time.Duration
    GetQualityScore() float64
    GetStats() TranslationStats
}

// TextPreprocessor handles text preprocessing
type TextPreprocessor interface {
    Clean(text string) string
    Tokenize(text string) ([]string, error)
    SegmentSentences(text string) []string
}

// TranslationEngine performs translation
type TranslationEngine interface {
    Initialize(modelPath string) error
    Translate(tokens []string, context []string) ([]Hypothesis, error)
    GetSupportedLanguages() []LanguagePair
}

// ContextManager manages translation context
type ContextManager interface {
    AddSentence(source, target string)
    GetContext() []SentencePair
    CacheTerm(source, target string)
    GetCachedTerm(source string) (string, bool)
    UpdateConversationState(topic string)
}

// QualityAssurance validates translation quality
type QualityAssurance interface {
    ValidateSemantic(source, target string) (float64, error)
    ScoreFluency(text string) (float64, error)
    NormalizeLength(source, target string) string
}

// PostProcessor handles post-processing
type PostProcessor interface {
    Detokenize(tokens []string) string
    AnnotateProsody(source, target string) ProsodyInfo
    AdjustFormatting(text string) string
}
```

## Data Models

```go
// TranslatedToken represents a translated text token
type TranslatedToken struct {
    SourceText    string
    TargetText    string
    SourceLang    string
    TargetLang    string
    Timestamp     time.Time
    Confidence    float64
    SemanticScore float64
    FluencyScore  float64
    ProsodyMarker ProsodyInfo
}

// ProsodyInfo contains prosodic annotations
type ProsodyInfo struct {
    RelativeDuration float64       // Speed adjustment (0.8-1.2)
    EmphasisLevel    int           // 0=none, 1=moderate, 2=strong
    PauseAfter       time.Duration // Pause duration after text
    Pitch            float64       // Pitch adjustment (-1.0 to 1.0)
}

// TranslationConfig holds configuration
type TranslationConfig struct {
    ModelPath        string
    SourceLanguage   string
    TargetLanguage   string
    BeamSize         int
    ContextWindow    int
    MinSemanticScore float64
    MinFluencyScore  float64
    MaxLengthRatio   float64
}

// Hypothesis represents a translation hypothesis
type Hypothesis struct {
    Text          string
    Tokens        []string
    Score         float64
    SemanticScore float64
    FluencyScore  float64
}

// TranslationStats contains performance statistics
type TranslationStats struct {
    TranslationLatency time.Duration
    AverageBLEU        float64
    AverageSemantic    float64
    AverageFluency     float64
    SentencesProcessed int64
    TokensTranslated   int64
    ErrorCount         int64
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*



### Property 1: Text Extraction Completeness
*For any* ASR token received, extracting text content should produce a non-empty string when the token contains recognized text.
**Validates: Requirements 1.1**

### Property 2: Text Cleaning Preserves Content
*For any* input text, cleaning and normalizing should preserve semantic content while removing extraneous whitespace and special characters.
**Validates: Requirements 1.2**

### Property 3: Tokenization Produces Valid Units
*For any* text input, tokenization should produce a sequence of subword units compatible with the translation model vocabulary.
**Validates: Requirements 1.3**

### Property 4: Sentence Segmentation Correctness
*For any* multi-sentence text, segmentation should produce individual sentences with proper boundaries.
**Validates: Requirements 1.4**

### Property 5: Input Processing Latency
*For any* ASR token, input processing (extraction, cleaning, tokenization, segmentation) should complete within 20 milliseconds.
**Validates: Requirements 1.5**

### Property 6: Translation Inference Produces Output
*For any* valid tokenized Portuguese text, model inference should generate English translation logits with valid dimensions.
**Validates: Requirements 2.2**

### Property 7: Beam Search Hypothesis Count
*For any* translation logits and beam size N, beam search should produce at most N translation hypotheses.
**Validates: Requirements 2.3**

### Property 8: Semantic Similarity Threshold
*For any* translation produced, the semantic similarity between source and target embeddings should be above 0.85.
**Validates: Requirements 2.5**

### Property 9: Context Window Size Maintenance
*For any* sequence of sentences processed, the context window should maintain exactly the configured number of previous sentences (e.g., 3).
**Validates: Requirements 3.1**

### Property 10: Context Usage in Translation
*For any* sentence translation, the context window should be accessed and used to inform translation decisions.
**Validates: Requirements 3.2**

### Property 11: Terminology Caching Consistency
*For any* domain-specific term encountered multiple times, the cached translation should be used consistently across all occurrences.
**Validates: Requirements 3.3, 3.4**

### Property 12: Conversation State Tracking
*For any* conversation processed, the conversation state (topic, context) should be maintained and updated as new sentences are processed.
**Validates: Requirements 3.5**

### Property 13: Semantic Validation Computation
*For any* translation generated, semantic similarity should be computed between source and target text using embedding models.
**Validates: Requirements 4.1**

### Property 14: Low Quality Translation Rejection
*For any* translation with semantic similarity below 0.75, the translation should be rejected and retried with alternative hypotheses.
**Validates: Requirements 4.2**

### Property 15: Fluency Score Computation
*For any* translation generated, a fluency score measuring naturalness should be computed and should be in the range [0.0, 1.0].
**Validates: Requirements 4.3**

### Property 16: Length Ratio Constraint
*For any* translation produced, the length should be within 20% of the source text length (ratio between 0.8 and 1.2).
**Validates: Requirements 4.4**

### Property 17: Quality Validation Latency
*For any* sentence translation, quality validation (semantic similarity, fluency, length) should complete within 30 milliseconds.
**Validates: Requirements 4.5**

### Property 18: Detokenization Produces Readable Text
*For any* sequence of subword tokens, detokenization should produce readable English text without artifacts.
**Validates: Requirements 5.1**

### Property 19: Prosody Marker Addition
*For any* detokenized text, prosody markers (emphasis, pauses, pitch) should be added to guide TTS synthesis.
**Validates: Requirements 5.2**

### Property 20: Prosody Transfer from Source
*For any* translation with prosody markers, the markers should be derived from prosodic characteristics of the source Portuguese text.
**Validates: Requirements 5.3**

### Property 21: English Formatting Conventions
*For any* final output text, punctuation and capitalization should follow English language conventions.
**Validates: Requirements 5.4**

### Property 22: Translated Token Completeness
*For any* translated token emitted, the token should contain all required fields: source text, target text, languages, timestamp, confidence, semantic score, fluency score, and prosody markers.
**Validates: Requirements 5.5, 6.2**

### Property 23: Token Emission for Completed Translations
*For any* completed translation, a Translated Token should be emitted containing the English text and metadata.
**Validates: Requirements 6.1**

### Property 24: Backpressure Prevents Overflow
*For any* situation where the token stream channel is full, the Translation Module should pause translation until channel capacity is available.
**Validates: Requirements 6.4**

### Property 25: Temporal Ordering Preservation
*For any* sequence of input ASR tokens, the output Translated Tokens should maintain the same temporal ordering.
**Validates: Requirements 6.5**

### Property 26: Latency Tracking Completeness
*For any* sentence translated, the translation latency should be recorded and available in metrics.
**Validates: Requirements 7.1**

### Property 27: Quality Metrics Recording
*For any* translation generated, BLEU score, semantic similarity, and fluency score should be recorded.
**Validates: Requirements 7.2, 7.3**

### Property 28: Counter Monotonicity
*For any* sequence of operations, counters for tokens translated, sentences processed, and errors should be monotonically increasing.
**Validates: Requirements 7.4, 7.5**

### Property 29: Malformed Input Handling
*For any* malformed input text, the Translation Module should log the error and emit an empty translation with zero confidence without crashing.
**Validates: Requirements 8.2**

### Property 30: Inference Failure Retry
*For any* inference failure, the Translation Module should retry with a simpler decoding strategy before failing completely.
**Validates: Requirements 8.3**

### Property 31: Low Similarity Retranslation
*For any* translation with semantic similarity below threshold, the Translation Module should attempt retranslation with alternative beam search parameters.
**Validates: Requirements 8.5**

### Property 32: Token Reception Latency
*For any* ASR token emitted by the ASR Module, the Translation Module should receive and begin processing within 10 milliseconds.
**Validates: Requirements 10.2**

### Property 33: Upstream Backpressure Application
*For any* situation where Translation processes tokens slower than ASR produces them, backpressure should be applied to prevent buffer overflow.
**Validates: Requirements 10.3**

### Property 34: Downstream Token Delivery
*For any* Translated Token emitted, the TTS Module should receive the token within 10 milliseconds.
**Validates: Requirements 10.4**

### Property 35: Downstream Backpressure Detection
*For any* situation where TTS processes tokens slower than Translation produces them, Translation should detect backpressure and pause accordingly.
**Validates: Requirements 10.5**

## Error Handling

### Error Categories
```go
type TranslationError struct {
    Code      ErrorCode
    Message   string
    Module    string
    Timestamp time.Time
    Context   map[string]interface{}
    Cause     error
}

type ErrorCode int

const (
    ErrModelLoadFailed ErrorCode = iota
    ErrInvalidInput
    ErrTokenizationFailed
    ErrInferenceFailed
    ErrLowSemanticSimilarity
    ErrLowFluency
    ErrChannelFull
    ErrTimeout
)
```

### Error Handling Strategies
1. **Initialization Errors**: Fail fast
2. **Processing Errors**: Log and emit empty translation
3. **Quality Errors**: Retry with alternative parameters
4. **Resource Errors**: Apply backpressure and wait

## Testing Strategy

### Unit Testing
- Test each sub-module independently
- Test with known input/output pairs
- Test edge cases and error conditions
- Target: >80% code coverage

### Property-Based Testing
- Test all 35 correctness properties
- Generate random inputs for comprehensive coverage
- Verify invariants hold across all inputs
- Run 100+ iterations per property

### Integration Testing
- Test ASR → Translation integration
- Test Translation → TTS integration
- Test end-to-end pipeline
- Verify latency and quality requirements

## Implementation Notes

### Recommended Technologies
- **Translation Model**: NLLB-200, M2M-100, or DeepL API
- **Embedding Model**: Sentence-BERT for semantic similarity
- **Fluency Scorer**: GPT-2 perplexity or custom model
- **Tokenizer**: SentencePiece or Hugging Face tokenizers

### Performance Optimization
- Batch multiple sentences for parallel translation
- Cache model weights in memory
- Use quantized models (INT8) for speed
- GPU acceleration for inference

### Quality Assurance
- Monitor BLEU scores in real-time
- Track semantic similarity distribution
- Alert on quality degradation
- Collect user feedback for improvement
