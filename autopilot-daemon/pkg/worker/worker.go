package worker

import (
	"sync"

	"github.com/IBM/autopilot/pkg/healthcheck"

	"k8s.io/klog/v2"
)

// worker is a function that processes tasks from the task queue.
// it runs in a separate goroutine and listens for tasks to process.
// it removes the task from the runningTasks map once completed.
func worker(c chan TaskType, sm *sync.Map) {
	for task := range c {
		switch task {
		case TaskPeriodicCheck:
			healthcheck.PeriodicCheck()
		case TaskInvasiveCheck:
			healthcheck.InvasiveCheck()
		default:
			klog.Errorf("Unknown task type: %v", task)
		}

		// mark the task as completed
		sm.Delete(task)
	}
}
