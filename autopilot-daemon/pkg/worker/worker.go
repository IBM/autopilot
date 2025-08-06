package worker

import "sync"

// worker is a function that processes tasks from the task queue.
// it runs in a separate goroutine and listens for tasks to process.
// it removes the task from the runningTasks map once completed.
func worker(c chan TaskType, sm *sync.Map) {
	for task := range c {
		switch task {
		case TaskPeriodicCheck:
			// Handle periodic check
		case TaskInvasiveCheck:
			// Handle invasive check
		default:
			// Handle unknown task type
		}

		// mark the task as completed
		sm.Delete(task)
	}
}
