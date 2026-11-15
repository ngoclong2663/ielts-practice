package usecases

import (
	"context"
	"net/http"

	"ielts-practice/app/modules/writing/dto"
	"ielts-practice/pkg/exceptions"
)

// AnswerPlan stores the learner's choices for the Step 0 questions; the view then reveals the answers.
func (uc *WritingUseCases) AnswerPlan(ctx context.Context, userID, attemptID string, req dto.PlanRequestDto) (dto.AttemptDto, error) {
	attempt, task, lesson, err := uc.loadAttempt(ctx, userID, attemptID)
	if err != nil {
		return dto.AttemptDto{}, err
	}
	questions := lesson.Content.Questions
	if len(req.Answers) != len(questions) {
		return dto.AttemptDto{}, appErr(http.StatusBadRequest, exceptions.CodeWritingInvalidStep, "Answer every question")
	}
	for i, a := range req.Answers {
		if a < 0 || a >= len(questions[i].OptionsVI) {
			return dto.AttemptDto{}, appErr(http.StatusBadRequest, exceptions.CodeWritingInvalidStep, "Unknown option")
		}
	}
	if err := uc.repo.SavePlanAnswers(ctx, attempt.ID, req.Answers); err != nil {
		return dto.AttemptDto{}, internalErr("Failed to save the answers", err)
	}
	attempt.PlanAnswers = req.Answers
	return uc.attemptView(ctx, attempt, task, lesson)
}
