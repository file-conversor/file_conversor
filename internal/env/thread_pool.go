// internal/env/thread_pool.go

package env

import (
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/file-conversor/file_conversor/internal/interfaces"
)

type ThreadPool struct {
	maxWorkers  int
	workersChan chan struct{}
	mu          sync.Mutex
	wg          *sync.WaitGroup
	errChans    []chan error
	finished    bool
}

func NewThreadPool(maxWorkers int) *ThreadPool {
	cpuCores := runtime.NumCPU()
	if maxWorkers <= 0 || maxWorkers > cpuCores*2 {
		maxWorkers = cpuCores * 2 // default to 2x CPU cores if invalid value provided
	}

	return &ThreadPool{
		maxWorkers:  maxWorkers,
		workersChan: make(chan struct{}, maxWorkers),
		wg:          &sync.WaitGroup{},
		finished:    false,
	}
}

func (tp *ThreadPool) AddTask(task func(interfaces.ProgressIncrement) error) {
	tp.workersChan <- struct{}{} // acquire a worker slot

	tp.mu.Lock()
	defer tp.mu.Unlock()

	if tp.finished {
		panic("cannot add task after Wait() has been called")
	}

	tp.wg.Add(1)
	errChan := make(chan error, 1) // buffered channel to avoid blocking
	tp.errChans = append(tp.errChans, errChan)

	go func() {
		var workErr error
		defer func() {
			// Recover from panic (avoid crashing the whole process)
			if r := recover(); r != nil {
				workErr = fmt.Errorf("panic in work thread: %v", r)
			}

			// report any error from the work function or panic to the error channel
			errChan <- workErr
			// release the worker slot
			<-tp.workersChan

			tp.wg.Done() // mark this task as done in the WaitGroup
		}()

		// execute the task and capture any error
		workErr = task(func(increment int64) {
			// noop progress increment function (no need to report progress from worker threads)
		})
	}()
}

func (tp *ThreadPool) Wait() error {
	// lock to prevent adding new tasks while waiting
	tp.mu.Lock()
	defer tp.mu.Unlock()

	tp.finished = true // mark as finished to prevent adding new tasks
	tp.wg.Wait()       // wait for all tasks to complete

	// collect errors from all tasks, if any, and return a combined error
	var errGrp error
	for _, errChan := range tp.errChans {
		// read the error from the channel
		if err := <-errChan; err != nil {
			errGrp = errors.Join(err, errGrp)
		}

		// close the error channel to prevent goroutine leaks
		close(errChan)
	}
	tp.errChans = nil     // clear errChans to release references and prevent memory leaks
	close(tp.workersChan) // close the workers channel to prevent goroutine leaks

	return errGrp
}
