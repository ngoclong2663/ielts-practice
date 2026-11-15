package domain

const (
	StatusPublished = "published"
	ModuleAcademic  = "academic"
)

// Task types per task number (Epic 5 §6.3).
var taskTypes = map[int]map[string]bool{
	1: {"line_graph": true, "bar_chart": true, "pie_chart": true, "table": true, "mixed": true, "map": true, "process": true},
	2: {"opinion": true, "discussion": true, "advantages_disadvantages": true, "problem_solution": true, "two_part": true},
}

func ValidTaskType(taskNo int, taskType string) bool { return taskTypes[taskNo][taskType] }

// WordsAndMinutes are the official length and suggested time for each task.
func WordsAndMinutes(taskNo int) (int, int) {
	if taskNo == 1 {
		return 150, 20
	}
	return 250, 40
}
