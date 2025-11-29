package pipeline

import (
	"fmt"
	"sync"
	"time"
)

// ProcessorFunc is a function that processes a chunk
type ProcessorFunc func(input interface{}) (interface{}, error)

// PipelineManager manages parallel processing of chunks
type PipelineManager struct {
	numWorkers int
	queueSize  int
	
	// Worker pool
	workers   []*worker
	jobQueue  chan *job
	started   bool
	mu        sync.RWMutex
	
	// Statistics
	chunksProcessed int64
	errors          int64
	startTime       time.Time
}

// job represents a processing job
type job struct {
	chunkID   int
	input     interface{}
	processor ProcessorFunc
	resultCh  chan *result
}

// result represents a processing result
type result struct {
	output interface{}
	err    error
}

// worker processes jobs from the queue
type worker struct {
	id       int
	jobQueue chan *job
	quit     chan bool
	wg       *sync.WaitGroup
}

// PipelineStats holds pipeline statistics
type PipelineStats struct {
	NumWorkers      int
	QueueSize       int
	ChunksProcessed int64
	Errors          int64
	Uptime          time.Duration
}

// NewPipelineManager creates a new pipeline manager
func NewPipelineManager(numWorkers, queueSize int) *PipelineManager {
	return &PipelineManager{
		numWorkers: numWorkers,
		queueSize:  queueSize,
		jobQueue:   make(chan *job, queueSize),
		workers:    make([]*worker, numWorkers),
	}
}

// Start starts the pipeline manager and worker pool
func (pm *PipelineManager) Start() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if pm.started {
		return fmt.Errorf("pipeline already started")
	}
	
	pm.startTime = time.Now()
	pm.started = true
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < pm.numWorkers; i++ {
		w := &worker{
			id:       i,
			jobQueue: pm.jobQueue,
			quit:     make(chan bool),
			wg:       &wg,
		}
		pm.workers[i] = w
		wg.Add(1)
		go w.start()
	}
	
	return nil
}

// Stop stops the pipeline manager and all workers
func (pm *PipelineManager) Stop() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if !pm.started {
		return nil
	}
	
	// Stop all workers
	for _, w := range pm.workers {
		if w != nil {
			close(w.quit)
		}
	}
	
	// Wait for workers to finish
	for _, w := range pm.workers {
		if w != nil {
			w.wg.Wait()
		}
	}
	
	pm.started = false
	return nil
}

// ProcessChunk processes a chunk with the given processor function
// Returns the result or an error
func (pm *PipelineManager) ProcessChunk(chunkID int, input interface{}, processor ProcessorFunc) (interface{}, error) {
	pm.mu.RLock()
	if !pm.started {
		pm.mu.RUnlock()
		return nil, fmt.Errorf("pipeline not started")
	}
	pm.mu.RUnlock()
	
	// Create result channel
	resultCh := make(chan *result, 1)
	
	// Create job
	j := &job{
		chunkID:   chunkID,
		input:     input,
		processor: processor,
		resultCh:  resultCh,
	}
	
	// Submit job to queue
	select {
	case pm.jobQueue <- j:
		// Job submitted successfully
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timeout submitting job to queue")
	}
	
	// Wait for result
	select {
	case res := <-resultCh:
		if res.err != nil {
			pm.mu.Lock()
			pm.errors++
			pm.mu.Unlock()
			return nil, res.err
		}
		
		pm.mu.Lock()
		pm.chunksProcessed++
		pm.mu.Unlock()
		
		return res.output, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for result")
	}
}

// GetStats returns pipeline statistics
func (pm *PipelineManager) GetStats() PipelineStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	return PipelineStats{
		NumWorkers:      pm.numWorkers,
		QueueSize:       pm.queueSize,
		ChunksProcessed: pm.chunksProcessed,
		Errors:          pm.errors,
		Uptime:          time.Since(pm.startTime),
	}
}

// worker methods

func (w *worker) start() {
	defer w.wg.Done()
	
	for {
		select {
		case job := <-w.jobQueue:
			w.processJob(job)
		case <-w.quit:
			return
		}
	}
}

func (w *worker) processJob(j *job) {
	// Recover from panics
	defer func() {
		if r := recover(); r != nil {
			j.resultCh <- &result{
				err: fmt.Errorf("worker panic: %v", r),
			}
		}
	}()
	
	// Process the job
	output, err := j.processor(j.input)
	
	// Send result
	j.resultCh <- &result{
		output: output,
		err:    err,
	}
}
