package worker

import (
	"sync"

	"k8s.io/klog/v2"
)

// WorkerPool manages a pool of go-routines that process tasks concurrently.
type WorkerPool struct {
	// runningTasks keeps track of tasks currently being processed
	runningTasks *sync.Map
	// taskQueue is a channel where tasks are submitted for processing
	taskQueue chan TaskType
}

// CreateWorkerPool initializes a new WorkerPool with a specified number of workers.
func CreateWorkerPool(numberOfWorkers int) *WorkerPool {
	syncMap := &sync.Map{}
	taskQueue := make(chan TaskType)

	// start the specified number of workers
	for i := 0; i < numberOfWorkers; i++ {
		go worker(taskQueue, syncMap) // start a worker goroutine, see worker.go for implementation
	}

	return &WorkerPool{
		runningTasks: syncMap,
		taskQueue:    taskQueue,
	}
}

// Submit adds a task to the worker pool for processing.
func (wp *WorkerPool) Submit(task TaskType) {
	// check if the task is running
	if _, exists := wp.runningTasks.Load(task); exists {
		klog.InfoS("Task already running, skipping submission", "task", task.String())
		return // task is already running, do not submit again
	}

	// mark the task as running, so it won't be submitted again
	// the worker will remove it from runningTasks when done
	wp.runningTasks.Store(task, struct{}{})
	wp.taskQueue <- task

	klog.InfoS("Task submitted to worker pool", "task", task.String())
}
