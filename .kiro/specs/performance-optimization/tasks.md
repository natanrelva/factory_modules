# Implementation Plan - Performance Optimization

## Phase 1: Translation Cache (TDD)

- [x] 1. Implement Translation Cache


  - Create `pkg/cache/translation_cache.go`
  - Implement LRU cache with map + doubly-linked list
  - _Requirements: 3.1, 3.2, 3.3_


- [ ]* 1.1 Write property test for cache consistency
  - **Property 1: Cache Consistency**
  - **Validates: Requirements 3.2**
  - Test that Get(key) after Set(key, value) returns value
  - Use testing/quick for random operations


- [ ]* 1.2 Write property test for LRU eviction
  - **Property 3: Cache LRU Eviction**
  - **Validates: Requirements 3.3**
  - Test that oldest entry is evicted when cache is full

  - Generate random sequences of Set operations

- [ ] 1.3 Implement cache persistence
  - Save cache to JSON file

  - Load cache from JSON file
  - _Requirements: 3.4_

- [ ]* 1.4 Write unit tests for persistence
  - Test Save() and Load() functions
  - Test error handling
  - _Requirements: 3.4_

## Phase 2: Silence Detection (TDD)

- [ ] 2. Implement Silence Detector
  - Create `pkg/silence/detector.go`
  - Implement energy calculation
  - Implement threshold-based detection
  - _Requirements: 4.1, 4.2_

- [ ]* 2.1 Write property test for energy calculation
  - **Property 5: Silence Detection Accuracy**
  - **Validates: Requirements 4.1**
  - Test that energy < 0.01 is classified as silence
  - Generate random audio samples

- [ ]* 2.2 Write property test for false positive rate
  - **Property 5: Silence Detection Accuracy**
  - **Validates: Requirements 4.5**
  - Test that false positive rate < 5%
  - Use real audio samples

- [ ] 2.3 Integrate silence detection in pipeline
  - Skip ASR when silence detected
  - Record silence skip metrics
  - _Requirements: 4.2, 4.3_

## Phase 3: Metrics Collection (TDD)

- [ ] 3. Implement Metrics Collector
  - Create `pkg/metrics/collector.go`
  - Implement latency recording
  - Implement cache hit/miss tracking
  - Implement silence skip tracking
  - _Requirements: 6.1, 6.2, 6.3, 6.4_

- [ ]* 3.1 Write property test for metrics accuracy
  - **Property 8: Metrics Accuracy**
  - **Validates: Requirements 6.1**
  - Test that all recorded latencies are preserved
  - Generate random metric sequences

- [ ] 3.2 Implement metrics aggregation
  - Calculate averages
  - Calculate percentiles (p50, p95, p99)
  - Calculate rates
  - _Requirements: 6.5_

- [ ]* 3.3 Write unit tests for aggregation
  - Test average calculation
  - Test percentile calculation
  - Test rate calculation
  - _Requirements: 6.5_

## Phase 4: Pipeline Manager (TDD)

- [ ] 4. Implement Pipeline Manager
  - Create `pkg/pipeline/manager.go`
  - Implement worker pool
  - Implement chunk queue
  - Implement result ordering
  - _Requirements: 2.1, 2.2, 2.3_

- [ ]* 4.1 Write property test for order preservation
  - **Property 2: Parallel Processing Order**
  - **Validates: Requirements 2.3**
  - Test that output order matches input order
  - Generate random chunk sequences

- [ ]* 4.2 Write property test for parallel throughput
  - **Property 10: Parallel Throughput Gain**
  - **Validates: Requirements 2.3**
  - Test that throughput increases with workers
  - Measure processing time with 1, 2, 3 workers

- [ ] 4.3 Implement error handling
  - Worker panic recovery
  - Queue timeout handling
  - Result timeout handling
  - _Requirements: 2.4_

- [ ]* 4.4 Write unit tests for error handling
  - Test worker recovery
  - Test timeout handling
  - _Requirements: 2.4_

## Phase 5: Integration (TDD)

- [ ] 5. Integrate cache into translation
  - Modify `pkg/translation-argos/translator.go`
  - Check cache before translating
  - Store result in cache after translating
  - _Requirements: 3.1, 3.2_

- [ ]* 5.1 Write integration test for cache hit rate
  - **Property 9: Cache Hit Rate**
  - **Validates: Requirements 3.2**
  - Test with 40% repeated phrases
  - Measure cache hit rate > 35%

- [ ] 5.2 Integrate silence detection
  - Modify `cmd/dubbing-mvp/main.go`
  - Check silence before ASR
  - Skip processing if silence
  - _Requirements: 4.2_

- [ ]* 5.3 Write integration test for silence skip
  - **Property 6: Silence Skip Efficiency**
  - **Validates: Requirements 4.2**
  - Test that ASR is skipped for silence
  - Measure time saved

- [ ] 5.4 Integrate metrics collection
  - Add metrics to all components
  - Record latencies
  - Record cache hits/misses
  - _Requirements: 6.1, 6.2, 6.3_

## Phase 6: Chunk Size Optimization (TDD)

- [ ] 6. Implement configurable chunk size
  - Add `--chunk-size` validation (1-5s)
  - Update capture to use configured size
  - _Requirements: 5.1_

- [ ]* 6.1 Write property test for chunk size
  - **Property 7: Chunk Size Configuration**
  - **Validates: Requirements 5.1**
  - Test that all sizes 1-5s work correctly
  - Generate random chunk sizes

- [ ] 6.2 Measure latency with different sizes
  - Test with 1s, 2s, 3s, 4s, 5s
  - Record latency for each
  - _Requirements: 5.2_

- [ ]* 6.3 Write performance test for latency
  - **Property 1: Latency Reduction**
  - **Validates: Requirements 1.1**
  - Test that latency < 6s
  - Measure with 2s chunks

## Phase 7: Performance Modes (TDD)

- [ ] 7. Implement performance modes
  - Add `--mode` flag (low-latency, balanced, quality)
  - Configure chunk size per mode
  - Configure workers per mode
  - _Requirements: 7.1, 7.2, 7.3, 7.4_

- [ ]* 7.1 Write unit tests for mode configuration
  - Test low-latency mode settings
  - Test balanced mode settings
  - Test quality mode settings
  - _Requirements: 7.1, 7.2, 7.3_

- [ ] 7.2 Implement mode switching
  - Apply settings when mode changes
  - Report expected latency
  - _Requirements: 7.4, 7.5_

## Phase 8: End-to-End Testing

- [ ] 8. Checkpoint - Ensure all tests pass
  - Run all unit tests
  - Run all property tests
  - Run all integration tests
  - Ensure all tests pass, ask the user if questions arise.

- [ ]* 8.1 Write end-to-end performance test
  - Test complete pipeline with real audio
  - Measure total latency
  - Validate < 6s target
  - _Requirements: 1.1_

- [ ]* 8.2 Write end-to-end cache test
  - Test with repeated phrases
  - Measure cache hit rate
  - Validate > 40% target
  - _Requirements: 3.2_

- [ ]* 8.3 Write end-to-end parallel test
  - Test with multiple chunks
  - Measure throughput gain
  - Validate > 30% improvement
  - _Requirements: 2.3_

## Phase 9: Documentation and Cleanup

- [ ] 9. Update documentation
  - Update README with performance info
  - Update CURRENT_STATUS with new metrics
  - Create PERFORMANCE_GUIDE.md
  - _Requirements: All_

- [ ] 9.1 Add performance monitoring
  - Add `--metrics` flag to show real-time metrics
  - Add `--stats` command to show aggregated stats
  - _Requirements: 6.5_

- [ ] 9.2 Final validation
  - Run all tests
  - Measure all metrics
  - Validate all targets met
  - Create performance report

## Phase 10: Final Checkpoint

- [ ] 10. Final Checkpoint - Validate all improvements
  - Ensure all tests pass
  - Validate latency < 6s
  - Validate cache hit rate > 40%
  - Validate throughput gain > 30%
  - Ensure all tests pass, ask the user if questions arise.

## Testing Summary

### Property-Based Tests (10 total)
1. Cache consistency
2. LRU eviction
3. Energy calculation
4. False positive rate
5. Metrics accuracy
6. Order preservation
7. Parallel throughput
8. Cache hit rate
9. Silence skip efficiency
10. Chunk size configuration

### Unit Tests (~20 total)
- Cache operations
- Persistence
- Silence detection
- Metrics aggregation
- Error handling
- Mode configuration

### Integration Tests (~5 total)
- Cache integration
- Silence integration
- Metrics integration
- End-to-end latency
- End-to-end throughput

### Performance Tests (~3 total)
- Latency benchmark
- Throughput benchmark
- Memory/CPU usage

**Total Tests: ~38**
**Target Coverage: > 80%**
