package entity

// Value objects sent to the AI coach (writing/ports/out.ICoach).

type LessonRequest struct {
	Task     *Task
	Position string
}

type CheckSentence struct {
	Prompt  string // what the learner saw: the Vietnamese sentence, idea note or English outline
	ModelEN string
	Answer  string
}

type CheckRequest struct {
	Task      *Task
	Paragraph string
	GoalVI    string
	Level     int
	Sentences []CheckSentence
}

type ReviewRequest struct {
	Task      *Task
	Position  string
	Text      string
	WordCount int
}
