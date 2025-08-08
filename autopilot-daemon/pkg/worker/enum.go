package worker

// TaskType represents the type of task to be processed by the worker pool.
// it is either a periodic check or an invasive check.
type TaskType int

const (
	TaskPeriodicCheck TaskType = iota + 1
	TaskInvasiveCheck
)

// String returns a string representation of the TaskType.
func (t TaskType) String() string {
	switch t {
	case TaskPeriodicCheck:
		return "Periodic Check"
	case TaskInvasiveCheck:
		return "Invasive Check"
	default:
		return "Unknown Task Type"
	}
}
