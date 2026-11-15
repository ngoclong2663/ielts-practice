package usecases

import (
	"context"
	"net/http"
	"strings"

	"ielts-practice/app/modules/writing/domain/rules"

	"ielts-practice/app/modules/writing/domain/entity"
	"ielts-practice/app/modules/writing/dto"
	"ielts-practice/pkg/exceptions"
)

// CheckParagraph sends the learner's sentences for one paragraph to the coach (one AI call) and stores
// the latest answers and feedback. Re-checking a paragraph counts again against the daily limit.
func (uc *WritingUseCases) CheckParagraph(ctx context.Context, userID, attemptID, key string, req dto.CheckRequestDto) (dto.AttemptDto, error) {
	attempt, task, lesson, err := uc.loadAttempt(ctx, userID, attemptID)
	if err != nil {
		return dto.AttemptDto{}, err
	}
	if attempt.Status != entity.StatusInProgress {
		return dto.AttemptDto{}, appErr(http.StatusConflict, exceptions.CodeWritingSubmitted, "This attempt is already submitted")
	}
	var paragraph *entity.LessonParagraph
	for i := range lesson.Content.Paragraphs {
		if lesson.Content.Paragraphs[i].Key == key {
			paragraph = &lesson.Content.Paragraphs[i]
		}
	}
	if attempt.Level >= rules.MaxLevel || paragraph == nil || len(req.Answers) != len(paragraph.Sentences) {
		return dto.AttemptDto{}, appErr(http.StatusBadRequest, exceptions.CodeWritingInvalidStep, "Unknown paragraph or wrong number of sentences")
	}
	answers := make([]string, len(req.Answers))
	empty := true
	for i, a := range req.Answers {
		answers[i] = strings.Join(strings.Fields(a), " ")
		empty = empty && answers[i] == ""
	}
	if empty {
		return dto.AttemptDto{}, appErr(http.StatusBadRequest, exceptions.CodeWritingTooShort, "Write at least one sentence")
	}
	if err := uc.checkLimit(ctx, userID, entity.KindCheck); err != nil {
		return dto.AttemptDto{}, err
	}

	sentences := make([]entity.CheckSentence, len(answers))
	for i, s := range paragraph.Sentences {
		sentences[i] = entity.CheckSentence{Prompt: promptFor(s, attempt.Level), ModelEN: s.ModelEN, Answer: answers[i]}
	}
	cctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()
	feedback, usage, err := uc.coach.CheckParagraph(cctx, entity.CheckRequest{
		Task: task, Paragraph: key, GoalVI: paragraph.GoalVI, Level: attempt.Level, Sentences: sentences,
	})
	if err != nil {
		return dto.AttemptDto{}, coachErr(err)
	}
	uc.record(ctx, userID, &attempt.ID, entity.KindCheck, usage)
	if err := uc.repo.UpsertParagraphCheck(ctx, attempt.ID, &entity.ParagraphCheck{Key: key, Answers: answers, Feedback: feedback}); err != nil {
		return dto.AttemptDto{}, internalErr("Failed to save the paragraph", err)
	}
	return uc.attemptView(ctx, attempt, task, lesson)
}
