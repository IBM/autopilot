package worker

func worker(c chan TaskType) {
	for task := range c {
		switch task {
		case TaskPeriodicCheck:
			// Handle periodic check
		case TaskInvasiveCheck:
			// Handle invasive check
		default:
			// Handle unknown task type
		}
	}
}
