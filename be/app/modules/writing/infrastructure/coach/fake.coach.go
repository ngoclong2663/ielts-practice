package coach

import (
	"context"
	"fmt"
	"strings"

	"ielts-practice/app/modules/writing/domain/rules"

	writingout "ielts-practice/app/modules/writing/ports/out"

	"ielts-practice/app/modules/writing/domain/entity"
)

// FakeCoach answers offline and deterministically (VOCAB_FAKE_PROVIDERS=true; never in production).
//   - Lessons: two sentences per paragraph.
//   - Checks: a sentence with "drink medicine" is a literal translation; under 4 words is wrong;
//     otherwise correct.
//   - Reviews: bands follow the word count.
type FakeCoach struct{}

var _ writingout.ICoach = FakeCoach{}

func (FakeCoach) Configured() bool { return true }

func (FakeCoach) WriteLesson(_ context.Context, req entity.LessonRequest) (entity.LessonContent, entity.AIUsage, error) {
	l := entity.LessonContent{
		IdeasVI: []string{"Ý chính thứ nhất", "Ý chính thứ hai", "Ý chính thứ ba"},
		Questions: []entity.PlanQuestion{
			{QuestionVI: "Đề bài yêu cầu gì?", OptionsVI: []string{"Mô tả và so sánh", "Kể một câu chuyện", "Viết thư"}, Answer: 0, ExplainVI: "Đề yêu cầu mô tả và so sánh."},
			{QuestionVI: "Bài viết cần ít nhất bao nhiêu từ?", OptionsVI: []string{"100", fmt.Sprint(req.Task.MinWords), "400"}, Answer: 1, ExplainVI: "Đây là số từ tối thiểu."},
		},
	}
	for _, key := range rules.ParagraphKeys(req.Task.TaskNo) {
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

func (FakeCoach) WriteVocab(_ context.Context, _ entity.VocabRequest) (entity.LessonVocab, entity.AIUsage, error) {
	item := func(term, pos, vi string, synonyms ...string) entity.VocabItem {
		return entity.VocabItem{Term: term, POS: pos, VI: vi, Synonyms: synonyms,
			ExampleEN: "This is an example sentence that uses " + term + ".", NoteVI: ""}
	}
	return entity.LessonVocab{Groups: []entity.VocabGroup{
		{Key: entity.VocabTopic, Items: []entity.VocabItem{
			item("workforce", "noun", "lực lượng lao động", "labour force", "workers"),
			item("manufacturing sector", "phrase", "ngành sản xuất", "industry"),
			item("job security", "phrase", "sự ổn định việc làm"),
		}},
		{Key: entity.VocabParaphrase, Items: []entity.VocabItem{
			item("show", "verb", "cho thấy", "illustrate", "compare"),
			item("the number of workers", "phrase", "số lượng công nhân", "employment", "the workforce"),
		}},
		{Key: entity.VocabTask, Items: []entity.VocabItem{
			item("rise steadily", "phrase", "tăng đều", "increase gradually", "climb"),
			item("in contrast", "phrase", "ngược lại", "by contrast", "whereas"),
		}},
	}}, entity.AIUsage{Model: "fake"}, nil
}

func (FakeCoach) CheckParagraph(_ context.Context, req entity.CheckRequest) ([]entity.SentenceFeedback, entity.AIUsage, error) {
	out := make([]entity.SentenceFeedback, len(req.Sentences))
	for i, s := range req.Sentences {
		answer := strings.TrimSpace(s.Answer)
		switch {
		case answer == "":
			out[i] = entity.SentenceFeedback{Verdict: "empty", ErrorTypes: []string{}, ExplainVI: "Bạn chưa viết câu này.", CorrectedEN: s.ModelEN}
		case strings.Contains(strings.ToLower(answer), "drink medicine"):
			out[i] = entity.SentenceFeedback{Verdict: "understandable", ErrorTypes: []string{"literal_translation"},
				ExplainVI: "\"Uống thuốc\" là \"take medicine\".", CorrectedEN: strings.ReplaceAll(answer, "drink medicine", "take medicine")}
		case rules.CountWords(answer) < 4:
			out[i] = entity.SentenceFeedback{Verdict: "wrong", ErrorTypes: []string{"missing_idea"}, ExplainVI: "Câu còn thiếu ý.", CorrectedEN: s.ModelEN}
		default:
			out[i] = entity.SentenceFeedback{Verdict: "correct", ErrorTypes: []string{}, ExplainVI: "Tốt lắm!", CorrectedEN: answer}
		}
	}
	return out, entity.AIUsage{Model: "fake"}, nil
}

func (FakeCoach) ReviewEssay(_ context.Context, req entity.ReviewRequest) (entity.EssayFeedback, entity.AIUsage, error) {
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
