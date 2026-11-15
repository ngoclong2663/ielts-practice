package usecases

import (
	"context"
	"strings"

	"ielts-practice/app/modules/writing/domain"
	"ielts-practice/app/modules/writing/domain/entity"
	"ielts-practice/app/modules/writing/dto"
)

func (uc *WritingUseCases) GetAttempt(ctx context.Context, userID, attemptID string) (dto.AttemptDto, error) {
	attempt, task, lesson, err := uc.loadAttempt(ctx, userID, attemptID)
	if err != nil {
		return dto.AttemptDto{}, err
	}
	return uc.attemptView(ctx, attempt, task, lesson)
}

// loadAttempt returns the user's attempt with its task and lesson (404 for another user's attempt).
func (uc *WritingUseCases) loadAttempt(ctx context.Context, userID, attemptID string) (*entity.Attempt, *entity.Task, *entity.Lesson, error) {
	attempt, err := uc.repo.FindAttempt(ctx, userID, attemptID)
	if err != nil {
		return nil, nil, nil, internalErr("Failed to load the attempt", err)
	}
	if attempt == nil {
		return nil, nil, nil, attemptNotFound()
	}
	task, err := uc.repo.FindTaskByID(ctx, attempt.TaskID)
	if err != nil {
		return nil, nil, nil, internalErr("Failed to load task", err)
	}
	if task == nil {
		return nil, nil, nil, taskNotFound()
	}
	lesson, err := uc.repo.FindLessonByID(ctx, attempt.LessonID)
	if err != nil || lesson == nil {
		return nil, nil, nil, internalErr("Failed to load the lesson", err)
	}
	return attempt, task, lesson, nil
}

// promptFor returns what the learner works from at their support level (Epic 5 §4).
func promptFor(s entity.LessonSentence, level int) string {
	switch level {
	case 1:
		return s.VI
	case 2:
		return s.NoteVI
	default:
		return s.OutlineEN
	}
}

// attemptView hides what the learner should not see yet: answers to plan questions until answered,
// model sentences until a paragraph is checked, the model answer until the attempt is submitted.
func (uc *WritingUseCases) attemptView(ctx context.Context, a *entity.Attempt, task *entity.Task, lesson *entity.Lesson) (dto.AttemptDto, error) {
	checks, err := uc.repo.ParagraphChecks(ctx, a.ID)
	if err != nil {
		return dto.AttemptDto{}, internalErr("Failed to load the attempt", err)
	}
	var submission *entity.Submission
	if a.Status == domain.StatusSubmitted {
		if submission, err = uc.repo.FindSubmission(ctx, a.ID); err != nil {
			return dto.AttemptDto{}, internalErr("Failed to load the submission", err)
		}
	}
	byKey := make(map[string]entity.ParagraphCheck, len(checks))
	for _, c := range checks {
		byKey[c.Key] = c
	}
	submitted := submission != nil
	c := lesson.Content

	out := dto.AttemptDto{
		ID: a.ID, Task: taskDto(task, domain.TaskProgress{}), Level: a.Level, Position: a.Position, Status: a.Status,
		IdeasVI: []string{}, Questions: make([]dto.PlanQuestionDto, 0, len(c.Questions)), PlanAnswers: a.PlanAnswers,
		Paragraphs: []dto.ParagraphDto{}, UsefulLanguage: []entity.Hint{}, CreatedAt: a.CreatedAt, SubmittedAt: a.SubmittedAt,
	}
	if a.Level < domain.MaxLevel {
		out.IdeasVI = c.IdeasVI
	}
	for _, q := range c.Questions {
		qd := dto.PlanQuestionDto{QuestionVI: q.QuestionVI, OptionsVI: q.OptionsVI}
		if a.PlanAnswers != nil {
			answer := q.Answer
			qd.Answer, qd.ExplainVI = &answer, q.ExplainVI
		}
		out.Questions = append(out.Questions, qd)
	}
	// Level 4 is free writing: no guided paragraphs.
	if a.Level < domain.MaxLevel {
		for _, p := range c.Paragraphs {
			check, checked := byKey[p.Key]
			pd := dto.ParagraphDto{
				Key: p.Key, GoalVI: p.GoalVI, Sentences: make([]dto.SentenceDto, 0, len(p.Sentences)),
				Answers: []string{}, Feedback: []entity.SentenceFeedback{}, CheckedCount: check.CheckedCount,
			}
			for _, s := range p.Sentences {
				sd := dto.SentenceDto{Prompt: promptFor(s, a.Level), Hints: []entity.Hint{}}
				if a.Level <= 2 {
					sd.Hints = s.Hints
				}
				if checked || submitted {
					sd.ModelEN = s.ModelEN
				}
				pd.Sentences = append(pd.Sentences, sd)
			}
			if checked {
				pd.Answers, pd.Feedback = check.Answers, check.Feedback
			}
			out.Paragraphs = append(out.Paragraphs, pd)
		}
	}
	if a.Level <= 2 || submitted {
		out.UsefulLanguage = c.UsefulLanguage
	}
	if submitted {
		out.ModelAnswer = modelAnswer(c)
		out.Submission = &dto.SubmissionDto{
			Text: submission.Text, WordCount: submission.WordCount, SecondsSpent: submission.SecondsSpent,
			BandLow: submission.BandLow, BandHigh: submission.BandHigh, Feedback: submission.Feedback, CreatedAt: submission.CreatedAt,
		}
	}
	return out, nil
}

// modelAnswer joins the lesson's model sentences into paragraphs.
func modelAnswer(c entity.LessonContent) string {
	paras := make([]string, 0, len(c.Paragraphs))
	for _, p := range c.Paragraphs {
		sentences := make([]string, 0, len(p.Sentences))
		for _, s := range p.Sentences {
			sentences = append(sentences, s.ModelEN)
		}
		paras = append(paras, strings.Join(sentences, " "))
	}
	return strings.Join(paras, "\n\n")
}
