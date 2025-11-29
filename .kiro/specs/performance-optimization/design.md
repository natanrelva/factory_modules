# Design Document - Performance Optimization

## Overview

Este documento descreve o design para otimização de performance do sistema de dublagem em tempo real. O objetivo é reduzir a latência de ~10s para ~5s através de:

1. **Processamento Paralelo** - Processar múltiplos chunks simultaneamente
2. **Cache de Traduções** - Reutilizar traduções de frases repetidas
3. **Detecção de Silêncio** - Pular processamento de áudio sem fala
4. **Chunks Menores** - Reduzir latência percebida
5. **Métricas Detalhadas** - Monitorar e identificar gargalos

## Architecture

### Current Architecture (Sequential)
```
┌─────────┐    ┌─────┐    ┌────────────┐    ┌─────┐    ┌────────┐
│ Capture │ -> │ ASR │ -> │ Translation│ -> │ TTS │ -> │ Output │
└─────────┘    └─────┘    └────────────┘    └─────┘    └────────┘
   ~3s          ~2s           ~4.5s          ~0.6s        ~0s
                        Total: ~10s
```

### New Architecture (Parallel + Cache)
```
┌─────────┐    ┌─────────────────────────────────────────┐
│ Capture │ -> │         Pipeline Manager                │
└─────────┘    │  ┌─────┐  ┌────────────┐  ┌─────┐     │
               │  │ ASR │  │ Translation│  │ TTS │     │
               │  └─────┘  └────────────┘  └─────┘     │
               │     ↓            ↓            ↓        │
               │  ┌─────────────────────────────────┐  │
               │  │      Translation Cache          │  │
               │  └─────────────────────────────────┘  │
               │  ┌─────────────────────────────────┐  │
               │  │      Silence Detector           │  │
               │  └─────────────────────────────────┘  │
               │  ┌─────────────────────────────────┐  │
               │  │      Metrics Collector          │  │
               │  └─────────────────────────────────┘  │
               └─────────────────────────────────────────┘
                                ↓
                          ┌────────┐
                          │ Output │
                          └────────┘
                    Target: ~5s
```

## Components and Interfaces

### 1. Pipeline Manager

**Responsabilidade**: Coordenar processamento paralelo de chunks

```go
type PipelineManager struct {
    workers      int
    chunkQueue   chan AudioChunk
    resultQueue  chan ProcessedChunk
    cache        *TranslationCache
    detector     *SilenceDetector
    metrics      *MetricsCollector
}

func (pm *PipelineManager) ProcessChunk(chunk AudioChunk) ProcessedChunk
func (pm *PipelineManager) Start() error
func (pm *PipelineManager) Stop() error
func (pm *PipelineManager) GetMetrics() Metrics
```

### 2. Translation Cache

**Responsabilidade**: Armazenar e recuperar traduções

```go
type TranslationCache struct {
    cache    map[string]CacheEntry
    maxSize  int
    lru      *LRUList
    mu       sync.RWMutex
}

type CacheEntry struct {
    Translation string
    Timestamp   time.Time
    HitCount    int
}

func (tc *TranslationCache) Get(text string) (string, bool)
func (tc *TranslationCache) Set(text, translation string)
func (tc *TranslationCache) GetStats() CacheStats
func (tc *TranslationCache) Save(path string) error
func (tc *TranslationCache) Load(path string) error
```

### 3. Silence Detector

**Responsabilidade**: Detectar silêncio em áudio

```go
type SilenceDetector struct {
    threshold float32
    minLength time.Duration
}

func (sd *SilenceDetector) IsSilence(samples []float32) bool
func (sd *SilenceDetector) GetEnergy(samples []float32) float32
func (sd *SilenceDetector) GetStats() SilenceStats
```

### 4. Metrics Collector

**Responsabilidade**: Coletar e agregar métricas

```go
type MetricsCollector struct {
    latencies    []ComponentLatency
    cacheHits    int64
    cacheMisses  int64
    silenceSkips int64
    mu           sync.RWMutex
}

type ComponentLatency struct {
    Component string
    Duration  time.Duration
    Timestamp time.Time
}

func (mc *MetricsCollector) RecordLatency(component string, duration time.Duration)
func (mc *MetricsCollector) RecordCacheHit()
func (mc *MetricsCollector) RecordCacheMiss()
func (mc *MetricsCollector) RecordSilenceSkip()
func (mc *MetricsCollector) GetAggregated() AggregatedMetrics
```

## Data Models

### AudioChunk
```go
type AudioChunk struct {
    ID        int
    Samples   []float32
    Timestamp time.Time
    Duration  time.Duration
}
```

### ProcessedChunk
```go
type ProcessedChunk struct {
    ID           int
    OriginalText string
    Translation  string
    AudioSamples []float32
    Latency      time.Duration
    CacheHit     bool
    WasSilence   bool
}
```

### Metrics
```go
type AggregatedMetrics struct {
    TotalChunks      int64
    AverageLatency   time.Duration
    CacheHitRate     float64
    SilenceSkipRate  float64
    ComponentLatencies map[string]time.Duration
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Latency Reduction
*For any* chunk processed with cache hit, the latency SHALL be less than 2 seconds
**Validates: Requirements 1.2**

### Property 2: Parallel Processing Order
*For any* sequence of chunks processed in parallel, the output order SHALL match the input order
**Validates: Requirements 2.3**

### Property 3: Cache Consistency
*For any* text that is cached, retrieving it SHALL return the same translation
**Validates: Requirements 3.2**

### Property 4: Cache LRU Eviction
*For any* cache at max capacity, adding a new entry SHALL remove the least recently used entry
**Validates: Requirements 3.3**

### Property 5: Silence Detection Accuracy
*For any* audio with energy < 0.01, the silence detector SHALL classify it as silence
**Validates: Requirements 4.1**

### Property 6: Silence Skip Efficiency
*For any* chunk classified as silence, ASR processing SHALL be skipped
**Validates: Requirements 4.2**

### Property 7: Chunk Size Configuration
*For any* chunk size between 1 and 5 seconds, the system SHALL process correctly
**Validates: Requirements 5.1**

### Property 8: Metrics Accuracy
*For any* chunk processed, all component latencies SHALL be recorded
**Validates: Requirements 6.1**

### Property 9: Cache Hit Rate
*For any* conversation with 40% repetition, cache hit rate SHALL be at least 35%
**Validates: Requirements 3.2**

### Property 10: Parallel Throughput Gain
*For any* workload processed with 3 workers, throughput SHALL increase by at least 25%
**Validates: Requirements 2.3**

## Error Handling

### Cache Errors
- **Cache Full**: Evict LRU entries
- **Cache Load Failure**: Start with empty cache
- **Cache Save Failure**: Log error, continue operation

### Parallel Processing Errors
- **Worker Panic**: Recover and restart worker
- **Queue Full**: Block until space available
- **Result Timeout**: Skip chunk and log error

### Silence Detection Errors
- **Invalid Audio**: Treat as non-silence
- **Threshold Misconfiguration**: Use default 0.01

## Testing Strategy

### Unit Tests
1. **Cache Operations**: Get, Set, Eviction, Persistence
2. **Silence Detection**: Energy calculation, threshold testing
3. **Metrics Collection**: Recording, aggregation, statistics
4. **Pipeline Manager**: Worker management, queue handling

### Property-Based Tests
Using **testing/quick** (Go's built-in PBT library):

1. **Property 1**: Cache consistency across random operations
2. **Property 2**: Parallel processing order preservation
3. **Property 3**: LRU eviction correctness
4. **Property 4**: Silence detection accuracy
5. **Property 5**: Metrics accuracy

### Integration Tests
1. **End-to-End Latency**: Measure with real audio
2. **Cache Hit Rate**: Test with repeated phrases
3. **Parallel Throughput**: Measure with multiple chunks
4. **Silence Skip**: Test with silent audio

### Performance Tests
1. **Latency Benchmark**: Target < 6s
2. **Throughput Benchmark**: Target > 0.5 chunks/s
3. **Memory Usage**: Target < 500 MB
4. **CPU Usage**: Target < 50%

## Implementation Plan

### Phase 1: Foundation (Week 1)
1. Implement Translation Cache with LRU
2. Implement Silence Detector
3. Implement Metrics Collector
4. Write unit tests for each component

### Phase 2: Integration (Week 2)
1. Implement Pipeline Manager
2. Integrate cache into translation flow
3. Integrate silence detection
4. Write integration tests

### Phase 3: Optimization (Week 3)
1. Implement parallel processing
2. Optimize chunk sizes
3. Add performance modes
4. Write performance tests

### Phase 4: Validation (Week 4)
1. Run all tests
2. Measure latency improvements
3. Validate cache hit rates
4. Document results

## Performance Targets

| Metric | Current | Target | Improvement |
|--------|---------|--------|-------------|
| Total Latency | ~10s | ~5s | 50% |
| Cache Hit Rate | 0% | 40% | N/A |
| Throughput | 0.33 chunks/s | 0.5 chunks/s | 50% |
| Silence Skip | 0% | 95% | N/A |
| CPU Usage | 30% | 40% | +33% |
| Memory Usage | 300 MB | 400 MB | +33% |

## Configuration

### Performance Modes

**Low-Latency Mode**:
```yaml
chunk_size: 2s
workers: 3
cache_enabled: true
silence_detection: true
```

**Balanced Mode** (Default):
```yaml
chunk_size: 3s
workers: 2
cache_enabled: true
silence_detection: true
```

**Quality Mode**:
```yaml
chunk_size: 4s
workers: 1
cache_enabled: false
silence_detection: false
```

## Monitoring

### Real-Time Metrics
- Current latency
- Cache hit rate
- Silence skip rate
- Active workers
- Queue sizes

### Aggregated Metrics
- Average latency (last 100 chunks)
- Component breakdown
- Cache statistics
- Error rates

## Rollback Plan

If performance degrades:
1. Disable parallel processing
2. Disable cache
3. Disable silence detection
4. Revert to sequential processing
5. Investigate and fix issues

## Success Criteria

1. ✅ Latency < 6s (target: 5s)
2. ✅ Cache hit rate > 40%
3. ✅ All tests passing
4. ✅ No regression in quality
5. ✅ Code coverage > 80%
6. ✅ Documentation complete
