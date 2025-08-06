package worker

// TaskType represents the type of task to be processed by the worker pool.
// it is either a periodic check or an invasive check.
type TaskType int

const (
	TaskPeriodicCheck TaskType = iota + 1
	TaskInvasiveCheck
)
