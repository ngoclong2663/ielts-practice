package live

import (
	"context"
	"encoding/json"
	writinginfra "ielts-practice/app/modules/writing/infrastructure"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	vocabinfra "ielts-practice/app/modules/vocab/infrastructure"
	"ielts-practice/app/modules/writing/domain"
	"ielts-practice/app/modules/writing/domain/entity"
	"ielts-practice/configs"
	"ielts-practice/seeds"
)

// Calls the real models (Claude first, Gemini fallback) for a Task 1 lesson, one paragraph check and one
// review; costs about a cent. Run with LLM_INTEGRATION=1; keys come from be/.env. WRITING_LIVE_OUT=<dir>
// writes the raw results there for reading.
func TestCoachLive(t *testing.T) {
	if os.Getenv("LLM_INTEGRATION") != "1" {
		t.Skip("set LLM_INTEGRATION=1 to call the real LLM APIs")
	}
	_ = godotenv.Load("../../.env") // be/.env
	cfg, err := configs.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.Vocab.FakeProviders = false
	coach := writinginfra.NewCoach(cfg, vocabinfra.NewJSONModel(cfg, zerolog.Nop()), zerolog.Nop())
	if !coach.Configured() {
		t.Skip("no API key")
	}
	slug, position := os.Getenv("WRITING_LIVE_TASK"), os.Getenv("WRITING_LIVE_POSITION")
	if slug == "" {
		slug = "cam21-test1-task1"
	}
	task := cam21Task(t, slug)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	out := map[string]any{}

	start := time.Now()
	lesson, usage, err := coach.WriteLesson(ctx, domain.LessonRequest{Task: task, Position: position})
	if err != nil {
		t.Fatalf("lesson: %v", err)
	}
	t.Logf("lesson by %s in %s, %d in / %d out tokens", usage.Model, time.Since(start).Round(time.Millisecond), usage.InputTokens, usage.OutputTokens)
	out["lesson"] = lesson
	words := 0
	for _, p := range lesson.Paragraphs {
		for _, s := range p.Sentences {
			words += domain.CountWords(s.ModelEN)
		}
	}
	t.Logf("model answer: %d words, %d questions, %d ideas", words, len(lesson.Questions), len(lesson.IdeasVI))

	intro := lesson.Paragraphs[0]
	sentences := make([]domain.CheckSentence, len(intro.Sentences))
	learner := []string{
		"The line graph show the number of worker in four industry in one country from 1965 to 2025.",
		"Overall, construction and fishing decrease, while tourism and education increase strongly.",
	}
	for i, s := range intro.Sentences {
		answer := ""
		if i < len(learner) {
			answer = learner[i]
		}
		sentences[i] = domain.CheckSentence{Prompt: s.VI, ModelEN: s.ModelEN, Answer: answer}
	}
	start = time.Now()
	feedback, usage, err := coach.CheckParagraph(ctx, domain.CheckRequest{Task: task, Paragraph: intro.Key, GoalVI: intro.GoalVI, Level: 1, Sentences: sentences})
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	t.Logf("check by %s in %s", usage.Model, time.Since(start).Round(time.Millisecond))
	out["check"] = feedback

	essay := `The line graph show how many people worked in four industries in one country from 1965 to 2025.

Overall, construction and fishing went down, while tourism and education increased a lot.

In 1965, construction had the most workers with 10 million, and it rose to 14 million in 1985. After that, it fell to 8 million in 2025. Fishing had 6 million workers at first but it decrease to only 2 million at the end.

Tourism started at 4 million and grew to 12.5 million in 2025. Education was the smallest in 1965, at 1.5 million, but it increase quickly and reached 13 million, which is the highest in 2025.`
	start = time.Now()
	review, usage, err := coach.ReviewEssay(ctx, domain.ReviewRequest{Task: task, Text: essay, WordCount: domain.CountWords(essay)})
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	t.Logf("review by %s in %s", usage.Model, time.Since(start).Round(time.Millisecond))
	out["review"] = review

	if dir := os.Getenv("WRITING_LIVE_OUT"); dir != "" {
		b, _ := json.MarshalIndent(out, "", "  ")
		if err := os.WriteFile(dir+"/coach-live.json", b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func cam21Task(t *testing.T, slug string) *entity.Task {
	t.Helper()
	raw, err := seeds.Writing.ReadFile("writing/cam21/pack.json")
	if err != nil {
		t.Fatal(err)
	}
	var pack struct {
		Tasks []struct {
			Slug     string          `json:"slug"`
			TaskNo   int             `json:"task_no"`
			TaskType string          `json:"task_type"`
			Title    string          `json:"title"`
			Prompt   string          `json:"prompt"`
			Facts    json.RawMessage `json:"facts"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal(raw, &pack); err != nil {
		t.Fatal(err)
	}
	for _, p := range pack.Tasks {
		if p.Slug == slug {
			minWords, minutes := domain.WordsAndMinutes(p.TaskNo)
			return &entity.Task{Slug: p.Slug, TaskNo: p.TaskNo, TaskType: p.TaskType, Title: p.Title,
				Prompt: strings.TrimSpace(p.Prompt), Facts: p.Facts, MinWords: minWords, Minutes: minutes}
		}
	}
	t.Fatalf("no task %s", slug)
	return nil
}
