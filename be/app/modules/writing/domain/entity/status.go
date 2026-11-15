package entity

// Values of Task.Status, Task.Module, Attempt.Status, and the kinds of AI calls (writing_ai_requests.kind).
const (
	StatusPublished = "published"
	ModuleAcademic  = "academic"

	StatusInProgress = "in_progress"
	StatusSubmitted  = "submitted"

	KindLesson = "lesson"
	KindCheck  = "check"
	KindSubmit = "submit"
)
