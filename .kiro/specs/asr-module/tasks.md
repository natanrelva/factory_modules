# Implementation Plan: ASR Module

## Overview

This implementation plan breaks down the ASR Module into discrete, actionable coding tasks. Each task builds incrementally on previous tasks, following the architecture established in the design document. The plan follows the successful patterns from the Audio Interface Module (M6) implementation.

---

## Tasks

- [ ] 1. Set up project structure and core interfaces
  - Create directory structure: `audio-interface/pkg/asr/` with subdirectories for each sub-module
  - Define core interfaces in `audio-interface/pkg/asr/interfaces/`
  - Define data types in `audio-interface/pkg/asr/types/`
  - Set up test infrastructure with Go testing framework
  - _Requirements: 9.1, 9.2_

- [ ] 2. Implement audio preprocessing sub-modules
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [ ] 2.1 Implement FeatureExtractor
  - Create `audio-interface/pkg/asr/preprocessing/feature_extractor.go`
  - Implement MFCC extraction using audio processing library
  - Implement Mel-spectrogram extraction
  - Add feature dimension configuration
  - _Requirements: 1.1_

- [ ]* 2.2 Write property test for feature extraction
  - **Property 1: Feature Extraction Produces Valid Dimensions**
  - **Validates: Requirements 1.1**
  - Generate random PCM frames with various properties
  - Extract features and verify dimensions match expected [time, feature_dim]
  - Test both MFCC and Mel-spectrogram feature types

- [ ] 2.3 Implement AudioNormalizer
  - Create `audio-interface/pkg/asr/preprocessing/audio_normalizer.go`
  - Implement amplitude normalization to [-1.0, 1.0] range
  - Implement DC offset removal
  - Add configurable normalization parameters
  - _Requirements: 1.2_

- [ ]* 2.4 Write property test for audio normalization
  - **Property 2: Audio Normalization Maintains Range**
  - **Validates: Requirements 1.2**
  - Generate random audio frames with various amplitudes
  - Normalize frames and verify all samples are in [-1.0, 1.0]
  - Test edge cases: silent audio, clipped audio, DC offset

- [ ] 2.5 Implement VADEnhancer
  - Create `audio-interface/pkg/asr/preprocessing/vad_enhancer.go`
  - Implement energy-based voice activity detection
  - Add configurable energy threshold
  - Implement silence detection with 500ms threshold
  - _Requirements: 1.3, 1.4_

- [ ]* 2.6 Write property test for VAD
  - **Property 3: VAD Correctly Classifies Voice Activity**
  - **Validates: Requirements 1.3**
  - Generate random frames with various energy levels
  - Apply VAD and verify classification consistency with threshold
  - Test boundary conditions around threshold

- [ ]* 2.7 Write unit tests for preprocessing modules
  - Test FeatureExtractor with known audio samples
  - Test AudioNormalizer with various amplitude levels
  - Test VADEnhancer with speech and silence samples
  - Test edge cases: empty frames, very short/long frames
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [ ] 3. Implement recognition engine sub-modules
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

- [ ] 3.1 Implement ModelLoader
  - Create `audio-interface/pkg/asr/recognition/model_loader.go`
  - Implement model loading from file path (support ONNX/PyTorch)
  - Extract and store model metadata (vocab size, sample rate, etc.)
  - Implement model validation and error handling
  - Add model unloading and cleanup
  - _Requirements: 2.1_

- [ ] 3.2 Implement InferenceEngine
  - Create `audio-interface/pkg/asr/recognition/inference_engine.go`
  - Implement forward pass through loaded model
  - Add batch processing support
  - Track inference timing for RTF calculation
  - Implement GPU support if available
  - _Requirements: 2.2, 2.5_

- [ ]* 3.3 Write property test for model inference
  - **Property 4: Model Inference Produces Valid Logits**
  - **Validates: Requirements 2.2**
  - Generate random feature inputs with valid dimensions
  - Execute inference and verify logit dimensions [time, vocab_size]
  - Verify vocab_size matches model configuration

- [ ]* 3.4 Write property test for RTF
  - **Property 6: Real-Time Factor Stays Below Threshold**
  - **Validates: Requirements 2.5**
  - Process random audio chunks of various lengths
  - Measure processing time vs audio duration
  - Verify RTF < 0.5 for all chunks

- [ ] 3.5 Implement BeamSearchDecoder
  - Create `audio-interface/pkg/asr/recognition/beam_decoder.go`
  - Implement beam search algorithm with configurable beam size
  - Generate multiple hypotheses with scores
  - Implement hypothesis ranking by combined score
  - _Requirements: 2.3_

- [ ]* 3.6 Write property test for beam search
  - **Property 5: Beam Search Produces Multiple Hypotheses**
  - **Validates: Requirements 2.3**
  - Generate random logits
  - Apply beam search with various beam sizes (1, 3, 5, 10)
  - Verify hypothesis count ≤ beam size
  - Verify hypotheses are ranked by score

- [ ]* 3.7 Write unit tests for recognition engine
  - Test ModelLoader with valid and invalid model paths
  - Test InferenceEngine with various feature inputs
  - Test BeamSearchDecoder with known logit sequences
  - Test error handling for model loading failures
  - _Requirements: 2.1, 2.2, 2.3_

- [ ] 4. Implement streaming management sub-modules
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [ ] 4.1 Implement ChunkManager
  - Create `audio-interface/pkg/asr/streaming/chunk_manager.go`
  - Implement audio chunking with configurable chunk size (1-3s)
  - Handle chunk boundaries and overlap
  - Implement frame buffering for chunk assembly
  - _Requirements: 3.1_

- [ ]* 4.2 Write property test for chunking
  - **Property 7: Chunking Produces Correct Sizes**
  - **Validates: Requirements 3.1**
  - Generate random continuous audio streams
  - Apply chunking with various configured sizes
  - Verify chunk durations are within [1s, 3s] range
  - Allow final chunk to be shorter

- [ ] 4.3 Implement ContextWindow
  - Create `audio-interface/pkg/asr/streaming/context_window.go`
  - Implement fixed-size sliding window for previous chunks
  - Add methods to add chunks and retrieve context
  - Implement context clearing
  - _Requirements: 3.2_

- [ ]* 4.4 Write property test for context window
  - **Property 8: Context Window Maintains Fixed Size**
  - **Validates: Requirements 3.2**
  - Process random sequences of chunks
  - Verify context window maintains exactly N previous chunks
  - Test window behavior when fewer than N chunks processed

- [ ] 4.5 Implement PartialHypothesis
  - Create `audio-interface/pkg/asr/streaming/partial_hypothesis.go`
  - Implement partial result emission with IsFinal flag
  - Create channel for streaming partial results
  - Implement timing control for partial emission (every 500ms)
  - _Requirements: 3.3, 3.4, 3.5_

- [ ]* 4.6 Write property test for partial hypotheses
  - **Property 9: Partial Hypotheses Precede Final Results**
  - **Property 10: Partial Hypothesis Flag Correctness**
  - **Validates: Requirements 3.3, 3.4, 3.5**
  - Process random chunks and emit partial/final results
  - Verify partial results are emitted before final results
  - Verify IsFinal=false for partials, IsFinal=true for finals
  - Verify final results emitted within 200ms of chunk completion

- [ ]* 4.7 Write unit tests for streaming management
  - Test ChunkManager with various audio lengths
  - Test ContextWindow with different window sizes
  - Test PartialHypothesis emission timing
  - Test edge cases: empty audio, single frame, very long audio
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [ ] 5. Implement post-processing sub-modules
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [ ] 5.1 Implement TextNormalizer
  - Create `audio-interface/pkg/asr/postprocessing/text_normalizer.go`
  - Implement Portuguese punctuation normalization rules
  - Implement capitalization rules (sentence start, proper nouns)
  - Implement whitespace normalization
  - _Requirements: 4.1, 4.2_

- [ ]* 5.2 Write property test for text normalization
  - **Property 11: Text Normalization Preserves Content**
  - **Validates: Requirements 4.1, 4.2**
  - Generate random recognized text with various formatting issues
  - Normalize text and verify semantic content is preserved
  - Verify punctuation and capitalization are improved

- [ ] 5.3 Implement TimestampAligner
  - Create `audio-interface/pkg/asr/postprocessing/timestamp_aligner.go`
  - Implement word-level timestamp alignment algorithm
  - Achieve 50ms alignment precision
  - Handle edge cases: single word, very long utterances
  - _Requirements: 4.3_

- [ ]* 5.4 Write property test for timestamp alignment
  - **Property 12: Timestamp Alignment Precision**
  - **Validates: Requirements 4.3**
  - Generate random text with known audio positions
  - Align timestamps and verify precision within 50ms
  - Test with various utterance lengths

- [ ] 5.5 Implement ConfidenceScorer
  - Create `audio-interface/pkg/asr/postprocessing/confidence_scorer.go`
  - Implement word-level confidence calculation from hypothesis scores
  - Implement utterance-level confidence aggregation (average)
  - Ensure confidence scores are in [0.0, 1.0] range
  - _Requirements: 4.4, 4.5_

- [ ]* 5.6 Write property test for confidence scoring
  - **Property 13: Confidence Scores Within Valid Range**
  - **Property 14: Aggregate Confidence Derived from Words**
  - **Validates: Requirements 4.4, 4.5**
  - Generate random hypotheses with various scores
  - Compute confidence scores and verify range [0.0, 1.0]
  - Verify aggregate confidence is derived from word confidences

- [ ]* 5.7 Write unit tests for post-processing
  - Test TextNormalizer with Portuguese text samples
  - Test TimestampAligner with known audio-text pairs
  - Test ConfidenceScorer with various hypothesis scores
  - Test edge cases: empty text, single word, perfect/zero confidence
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [ ] 6. Implement language model integration sub-modules
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [ ] 6.1 Implement LanguageModelLoader
  - Create `audio-interface/pkg/asr/lm/lm_loader.go`
  - Implement language model loading (support KenLM format)
  - Add model validation and metadata extraction
  - Implement model unloading and cleanup
  - _Requirements: 5.1_

- [ ] 6.2 Implement LMRescorer
  - Create `audio-interface/pkg/asr/lm/lm_rescorer.go`
  - Implement hypothesis rescoring with language model
  - Combine acoustic and language model scores with configurable weight
  - Implement hypothesis reranking by combined score
  - Ensure rescoring completes within 30ms per hypothesis
  - _Requirements: 5.2, 5.3, 5.5_

- [ ]* 6.3 Write property test for LM rescoring
  - **Property 15: Language Model Rescoring Changes Scores**
  - **Property 16: Best Hypothesis Selection**
  - **Validates: Requirements 5.2, 5.3**
  - Generate random hypothesis sets
  - Apply LM rescoring and verify scores are updated
  - Verify selected hypothesis has highest combined score
  - Verify rescoring completes within 30ms

- [ ]* 6.4 Write unit tests for language model integration
  - Test LanguageModelLoader with valid and invalid LM paths
  - Test LMRescorer with various hypothesis sets
  - Test LM weight adjustment
  - Test error handling for LM loading failures
  - _Requirements: 5.1, 5.2, 5.3, 5.5_

- [ ] 7. Implement token stream output
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ] 7.1 Implement ASRToken creation and emission
  - Create `audio-interface/pkg/asr/types/token.go`
  - Implement ASRToken struct with all required fields
  - Implement token validation
  - Create token stream channel with buffering
  - _Requirements: 6.1, 6.2, 6.3_

- [ ]* 7.2 Write property test for token completeness
  - **Property 17: Token Completeness**
  - **Property 18: Word Timings Match Text**
  - **Validates: Requirements 6.2, 6.3**
  - Generate random ASR tokens
  - Verify all required fields are present and valid
  - Verify word timing count equals word count in text
  - Verify each word timing corresponds to a word

- [ ] 7.3 Implement backpressure handling
  - Add backpressure logic when token channel is full
  - Pause audio processing when backpressure is active
  - Resume processing when channel has capacity
  - _Requirements: 6.5_

- [ ]* 7.4 Write property test for backpressure
  - **Property 19: Backpressure Prevents Overflow**
  - **Validates: Requirements 6.5**
  - Fill token stream channel to capacity
  - Attempt to emit more tokens
  - Verify processing pauses until capacity available
  - Verify no tokens are lost

- [ ]* 7.5 Write unit tests for token stream
  - Test ASRToken creation with various inputs
  - Test token validation
  - Test channel communication
  - Test backpressure behavior
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ] 8. Implement observability sub-modules
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6_

- [ ] 8.1 Implement LatencyTracker
  - Create `audio-interface/pkg/asr/metrics/latency_tracker.go`
  - Track processing latency for each chunk
  - Calculate average, P50, P95, P99 latencies
  - Implement sliding window for recent latencies
  - _Requirements: 7.1_

- [ ]* 8.2 Write property test for latency tracking
  - **Property 20: Latency Tracking Completeness**
  - **Property 21: RTF Calculation Correctness**
  - **Validates: Requirements 7.1, 7.2**
  - Process random chunks and record latencies
  - Verify all latencies are recorded
  - Verify RTF calculation: processing_time / audio_duration
  - Verify RTF is always positive

- [ ] 8.3 Implement AccuracyMonitor
  - Create `audio-interface/pkg/asr/metrics/accuracy_monitor.go`
  - Estimate WER based on confidence scores
  - Track average confidence over time
  - Implement sliding window for recent accuracy metrics
  - _Requirements: 7.3_

- [ ] 8.4 Implement MetricsCollector
  - Create `audio-interface/pkg/asr/metrics/collector.go`
  - Aggregate metrics from all sub-modules
  - Implement ASRStats struct population
  - Track counters: frames processed, tokens emitted, errors
  - Provide metrics query interface
  - _Requirements: 7.4, 7.5, 7.6_

- [ ]* 8.5 Write property test for counter monotonicity
  - **Property 22: Counter Monotonicity**
  - **Validates: Requirements 7.4, 7.5**
  - Process random sequences of operations
  - Verify counters (frames, tokens, errors) never decrease
  - Verify counters increment correctly for each operation

- [ ]* 8.6 Write unit tests for observability
  - Test LatencyTracker with various latency values
  - Test AccuracyMonitor with various confidence scores
  - Test MetricsCollector aggregation
  - Test percentile calculations
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6_

- [ ] 9. Implement error handling and resilience
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [ ] 9.1 Implement error types and error handling
  - Create `audio-interface/pkg/asr/errors.go`
  - Define ASRError struct with error codes
  - Implement error wrapping and context
  - Add error logging with appropriate levels
  - _Requirements: 8.1, 8.2, 8.3_

- [ ] 9.2 Implement error recovery logic
  - Add corrupted frame handling (log and continue)
  - Add inference failure handling (emit empty token, continue)
  - Add memory exhaustion recovery (clear buffers, attempt recovery)
  - Implement degraded state after 3 consecutive failures
  - _Requirements: 8.2, 8.3, 8.4, 8.5_

- [ ]* 9.3 Write property test for error recovery
  - **Property 23: Error Recovery Continuity**
  - **Validates: Requirements 8.2, 8.3**
  - Inject random errors during processing
  - Verify processing continues after errors
  - Verify error counters increment correctly
  - Verify no crashes or panics

- [ ]* 9.4 Write unit tests for error handling
  - Test initialization error handling
  - Test processing error handling
  - Test resource error handling
  - Test degraded state transitions
  - Test error logging and context
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [ ] 10. Implement ASRCoordinator (orchestration)
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

- [ ] 10.1 Implement ASRCoordinator structure
  - Create `audio-interface/pkg/asr/coordinator.go`
  - Define coordinator struct with all sub-module references
  - Implement configuration management
  - Add worker pool for parallel processing
  - _Requirements: 9.1_

- [ ] 10.2 Implement lifecycle management
  - Implement Initialize() method (load models, allocate resources)
  - Implement Start() method (begin processing)
  - Implement Stop() method (graceful shutdown)
  - Implement Close() method (cleanup resources)
  - _Requirements: 9.2, 9.3, 9.4, 9.5_

- [ ] 10.3 Implement core processing pipeline
  - Implement ProcessFrame() method
  - Orchestrate flow: preprocessing → recognition → post-processing
  - Implement chunk-based processing with context
  - Emit tokens through output channel
  - _Requirements: 1.1, 2.2, 3.1, 4.1, 6.1_

- [ ] 10.4 Implement configuration methods
  - Implement SetLanguage() method
  - Implement SetModel() method
  - Add runtime configuration updates
  - _Requirements: 9.1_

- [ ] 10.5 Implement observability methods
  - Implement GetLatency() method
  - Implement GetAccuracy() method
  - Implement GetStats() method
  - Aggregate metrics from all sub-modules
  - _Requirements: 7.6_

- [ ]* 10.6 Write unit tests for ASRCoordinator
  - Test coordinator initialization
  - Test lifecycle methods (Initialize, Start, Stop, Close)
  - Test configuration methods
  - Test observability methods
  - Test error propagation
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

- [ ] 11. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 12. Implement integration with Audio Interface Module
  - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_

- [ ] 12.1 Implement Audio Interface integration
  - Create integration layer in `audio-interface/pkg/asr/integration/`
  - Subscribe to PCM frame stream from Audio Interface Module
  - Implement frame reception with <10ms latency
  - Handle capture start/stop events
  - Query capture latency from Audio Interface Module
  - _Requirements: 10.1, 10.2, 10.5_

- [ ]* 12.2 Write property test for frame reception latency
  - **Property 24: Frame Reception Latency**
  - **Validates: Requirements 10.2**
  - Emit random PCM frames from Audio Interface Module
  - Measure time from emission to ASR processing start
  - Verify latency < 10ms for all frames

- [ ] 12.3 Implement backpressure coordination
  - Implement backpressure signaling to Audio Interface Module
  - Coordinate buffer levels between modules
  - Prevent frame loss during slow processing
  - _Requirements: 10.3_

- [ ]* 12.4 Write property test for backpressure coordination
  - **Property 25: Backpressure Prevents Buffer Overflow**
  - **Validates: Requirements 10.3**
  - Produce frames faster than ASR can process
  - Verify backpressure is applied
  - Verify no frames are lost
  - Verify processing resumes when capacity available

- [ ]* 12.5 Write integration tests for Audio Interface
  - Test end-to-end flow: Audio Interface → ASR
  - Test with various frame rates
  - Test event handling (start/stop)
  - Test backpressure scenarios
  - Test latency requirements
  - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_

- [ ] 13. Create example application and documentation
  - _Requirements: All_

- [ ] 13.1 Create example application
  - Create `audio-interface/cmd/asr-demo/main.go`
  - Demonstrate ASR module usage
  - Show integration with Audio Interface Module
  - Include configuration examples
  - Add performance monitoring

- [ ] 13.2 Write module documentation
  - Create `audio-interface/pkg/asr/README.md`
  - Document module architecture
  - Document API usage with examples
  - Document configuration options
  - Document performance characteristics

- [ ] 13.3 Write usage guide
  - Create `audio-interface/pkg/asr/USAGE_GUIDE.md`
  - Provide step-by-step setup instructions
  - Include model download and installation
  - Document common use cases
  - Include troubleshooting section

- [ ] 13.4 Write architecture documentation
  - Create `audio-interface/pkg/asr/ARCHITECTURE.md`
  - Document module decomposition
  - Document data flow
  - Document design decisions
  - Include diagrams (Mermaid)

- [ ] 14. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

---

## Summary

**Total Tasks**: 14 top-level tasks, 60+ sub-tasks
**Estimated Timeline**: 2-3 weeks
**Test Coverage Goal**: >80% line coverage
**Property Tests**: 25 correctness properties
**Integration Points**: Audio Interface Module (M6), Translation Module (M3)

**Key Milestones**:
1. Week 1: Audio preprocessing + Recognition engine (Tasks 1-3)
2. Week 2: Streaming + Post-processing + LM integration (Tasks 4-6)
3. Week 3: Token output + Observability + Integration (Tasks 7-12)
4. Final: Documentation and testing (Tasks 13-14)
