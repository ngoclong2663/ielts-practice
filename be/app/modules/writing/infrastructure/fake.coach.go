package infrastructure

import (
	"context"
	"fmt"
	"strings"

	"ielts-practice/app/modules/writing/domain"
	"ielts-practice/app/modules/writing/domain/entity"
)

// FakeCoach answers offline and deterministically (VOCAB_FAKE_PROVIDERS=true; never in production).
//   - Lessons: two sentences per paragraph.
//   - Checks: a sentence with "drink medicine" is a literal translation; under 4 words is wrong;
//     otherwise correct.
//   - Reviews: bands follow the word count.
type FakeCoach struct{}

var _ domain.ICoach = FakeCoach{}

func (FakeCoach) Configured() bool { return true }

func (FakeCoach) WriteLesson(_ context.Context, req domain.LessonRequest) (entity.LessonContent, entity.AIUsage, error) {
	l := entity.LessonContent{
		IdeasVI: []string{"Ý chính thứ nhất", "Ý chính thứ hai", "Ý chính thứ ba"},
		Questions: []entity.PlanQuestion{
			{QuestionVI: "Đề bài yêu cầu gì?", OptionsVI: []string{"Mô tả và so sánh", "Kể một câu chuyện", "Viết thư"}, Answer: 0, ExplainVI: "Đề yêu cầu mô tả và so sánh."},
			{QuestionVI: "Bài viết cần ít nhất bao nhiêu từ?", OptionsVI: []string{"100", fmt.Sprint(req.Task.MinWords), "400"}, Answer: 1, ExplainVI: "Đây là số từ tối thiểu."},
		},
		UsefulLanguage: []entity.Hint{{EN: "rise steadily", VI: "tăng đều"}, {EN: "in contrast", VI: "ngược lại"}},
	}
	for _, key := range domain.ParagraphKeys(req.Task.TaskNo) {
		p := entity.LessonParagraph{Key: key, GoalVI: "Mục tiêu của đoạn " + key}
		for i := 1; i <= 2; i++ {
			p.Sentences = append(p.Sentences, entity.LessonSentence{
				VI:        fmt.Sprintf("Câu %d của đoạn %s.", i, key),
				NoteVI:    fmt.Sprintf("ý %d", i),
				OutlineEN: fmt.Sprintf("%s / sentence %d", key, i),
				Hints:     []entity.Hint{{EN: "take medicine", VI: "uống thuốc"}},
				ModelEN:   fmt.Sprintf("This is model sentence %d of the %s paragraph.", i, key),
			})
		}
		l.Paragraphs = append(l.Paragraphs, p)
	}
	return l, entity.AIUsage{Model: "fake"}, nil
}

func (FakeCoach) CheckParagraph(_ context.Context, req domain.CheckRequest) ([]entity.SentenceFeedback, entity.AIUsage, error) {
	out := make([]entity.SentenceFeedback, len(req.Sentences))
	for i, s := range req.Sentences {
		answer := strings.TrimSpace(s.Answer)
		switch {
		case answer == "":
			out[i] = entity.SentenceFeedback{Verdict: "empty", ErrorTypes: []string{}, ExplainVI: "Bạn chưa viết câu này.", CorrectedEN: s.ModelEN}
		case strings.Contains(strings.ToLower(answer), "drink medicine"):
			out[i] = entity.SentenceFeedback{Verdict: "understandable", ErrorTypes: []string{"literal_translation"},
				ExplainVI: "\"Uống thuốc\" là \"take medicine\".", CorrectedEN: strings.ReplaceAll(answer, "drink medicine", "take medicine")}
		case domain.CountWords(answer) < 4:
			out[i] = entity.SentenceFeedback{Verdict: "wrong", ErrorTypes: []string{"missing_idea"}, ExplainVI: "Câu còn thiếu ý.", CorrectedEN: s.ModelEN}
		default:
			out[i] = entity.SentenceFeedback{Verdict: "correct", ErrorTypes: []string{}, ExplainVI: "Tốt lắm!", CorrectedEN: answer}
		}
	}
	return out, entity.AIUsage{Model: "fake"}, nil
}

func (FakeCoach) ReviewEssay(_ context.Context, req domain.ReviewRequest) (entity.EssayFeedback, entity.AIUsage, error) {
	band := 5.0
	if req.WordCount >= req.Task.MinWords {
		band = 6
	}
	fb := entity.EssayFeedback{SummaryVI: "Bài viết rõ ràng; hãy thêm ví dụ cụ thể.", Model: "fake",
		Fixes: []entity.Fix{{Quote: strings.Fields(req.Text)[0], Better: "Better", ExplainVI: "Ví dụ sửa lỗi."}}}
	for _, key := range criteriaKeys(req.Task.TaskNo) {
		fb.Criteria = append(fb.Criteria, entity.CriterionFeedback{Key: key, Band: band, CommentVI: "Nhận xét cho " + key})
	}
	return fb, entity.AIUsage{Model: "fake"}, nil
}
