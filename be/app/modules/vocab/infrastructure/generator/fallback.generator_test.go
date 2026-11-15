package generator

import (
	"context"
	"errors"
	"testing"

	vocabout "ielts-practice/app/modules/vocab/ports/out"

	"github.com/rs/zerolog"

	"ielts-practice/app/modules/vocab/domain/entity"
)

// stubGenerator answers the words in `answers` and returns err.
type stubGenerator struct {
	configured bool
	answers    map[string]bool
	err        error
	calls      [][]string
}

func (s *stubGenerator) Configured() bool { return s.configured }

func (s *stubGenerator) Generate(_ context.Context, words []entity.WordRequest) ([]entity.WordResult, entity.Usage, error) {
	var asked []string
	var out []entity.WordResult
	for _, w := range words {
		asked = append(asked, w.Input)
		if s.answers == nil || s.answers[w.Input] {
			out = append(out, entity.WordResult{Input: w.Input, IsValid: true})
		}
	}
	s.calls = append(s.calls, asked)
	return out, entity.Usage{InputTokens: 10 * len(words), OutputTokens: 100 * len(words)}, s.err
}

func requests(words ...string) []entity.WordRequest {
	out := make([]entity.WordRequest, len(words))
	for i, w := range words {
		out[i] = entity.WordRequest{Input: w}
	}
	return out
}

func TestFallbackGenerator(t *testing.T) {
	ctx := context.Background()
	log := zerolog.Nop()

	t.Run("primary answers everything", func(t *testing.T) {
		primary, fallback := &stubGenerator{configured: true}, &stubGenerator{configured: true}
		res, _, err := NewFallbackGenerator(primary, "p", fallback, "f", log).Generate(ctx, requests("a", "b"))
		if err != nil || len(res) != 2 || len(fallback.calls) != 0 {
			t.Fatalf("fallback should not be called: %v %d %v", err, len(res), fallback.calls)
		}
	})

	t.Run("primary fails: fallback answers all", func(t *testing.T) {
		primary := &stubGenerator{configured: true, answers: map[string]bool{}, err: errors.New("overloaded")}
		fallback := &stubGenerator{configured: true}
		res, usage, err := NewFallbackGenerator(primary, "p", fallback, "f", log).Generate(ctx, requests("a", "b"))
		if err != nil || len(res) != 2 || len(fallback.calls) != 1 || usage.InputTokens != 40 {
			t.Fatalf("unexpected: %v %d %v %+v", err, len(res), fallback.calls, usage)
		}
	})

	t.Run("only missing words go to the fallback", func(t *testing.T) {
		primary := &stubGenerator{configured: true, answers: map[string]bool{"a": true}}
		fallback := &stubGenerator{configured: true}
		res, _, err := NewFallbackGenerator(primary, "p", fallback, "f", log).Generate(ctx, requests("a", "b", "c"))
		if err != nil || len(res) != 3 || len(fallback.calls[0]) != 2 || fallback.calls[0][0] != "b" {
			t.Fatalf("unexpected: %v %d %v", err, len(res), fallback.calls)
		}
	})

	t.Run("no primary key: fallback only", func(t *testing.T) {
		primary, fallback := &stubGenerator{}, &stubGenerator{configured: true}
		g := NewFallbackGenerator(primary, "p", fallback, "f", log)
		res, _, err := g.Generate(ctx, requests("a"))
		if !g.Configured() || err != nil || len(res) != 1 || len(primary.calls) != 0 {
			t.Fatalf("unexpected: %v %d %v", err, len(res), primary.calls)
		}
	})

	t.Run("no fallback key: primary error is returned", func(t *testing.T) {
		primary := &stubGenerator{configured: true, answers: map[string]bool{}, err: vocabout.ErrGeneratorRefused}
		_, _, err := NewFallbackGenerator(primary, "p", &stubGenerator{}, "f", log).Generate(ctx, requests("a"))
		if !errors.Is(err, vocabout.ErrGeneratorRefused) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("both fail: fallback error is returned", func(t *testing.T) {
		primary := &stubGenerator{configured: true, answers: map[string]bool{}, err: errors.New("down")}
		fallback := &stubGenerator{configured: true, answers: map[string]bool{}, err: errors.New("also down")}
		res, _, err := NewFallbackGenerator(primary, "p", fallback, "f", log).Generate(ctx, requests("a"))
		if err == nil || len(res) != 0 {
			t.Fatalf("unexpected: %v %d", err, len(res))
		}
	})
}
