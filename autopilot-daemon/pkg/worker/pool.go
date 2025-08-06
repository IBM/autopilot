package worker

// WorkerPool manages a pool of go-routines that process tasks concurrently.
type WorkerPool struct {
	taskChannel chan TaskType
}

// CreateWorkerPool initializes a new WorkerPool with a specified number of workers.
func CreateWorkerPool(numberOfWorkers int) *WorkerPool {
	taskChannel := make(chan TaskType)

	for i := 0; i < numberOfWorkers; i++ {
		go worker(taskChannel)
	}

	return &WorkerPool{
		taskChannel: taskChannel,
	}
}

// Submit adds a task to the worker pool for processing.
func (wp *WorkerPool) Submit(task TaskType) {
	wp.taskChannel <- task
}
