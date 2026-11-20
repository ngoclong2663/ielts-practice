package usecases

import (
	"context"
	"slices"

	"ielts-practice/app/modules/exam/domain/entity"
	"ielts-practice/app/modules/exam/dto"
)

// statsWindow is how many recent attempts the accuracy per question type covers.
const statsWindow = 20

// GetStats returns accuracy per question type over the last 20 attempts (redo runs left out), weakest
// first, and the band of every full test, oldest first.
func (uc *ExamUseCases) GetStats(ctx context.Context, userID, skill string) (*dto.StatsDto, error) {
	attempts, err := uc.repo.ListAttempts(ctx, userID, skill, 0, "", true)
	if err != nil {
		return nil, internal("Failed to load attempts", err)
	}
	out := &dto.StatsDto{ByType: []dto.TypeStatDto{}, Bands: []dto.BandPointDto{}}
	byType := map[string]*dto.TypeStatDto{}
	var order []string
	for _, a := range attempts {
		if a.Kind == entity.KindRedo {
			continue
		}
		if a.Kind == entity.KindFull {
			out.Bands = append(out.Bands, dto.BandPointDto{AttemptID: a.ID, Band: a.Band, Correct: a.Correct, At: a.CreatedAt})
		}
		if out.Attempts >= statsWindow {
			continue
		}
		out.Attempts++
		for _, r := range a.Results {
			s, ok := byType[r.Type]
			if !ok {
				s = &dto.TypeStatDto{Type: r.Type}
				byType[r.Type] = s
				order = append(order, r.Type)
			}
			s.Total++
			if r.Correct {
				s.Correct++
			}
		}
	}
	for _, t := range order {
		out.ByType = append(out.ByType, *byType[t])
	}
	slices.SortStableFunc(out.ByType, func(a, b dto.TypeStatDto) int {
		// Compare a.Correct/a.Total with b.Correct/b.Total without floats.
		return a.Correct*b.Total - b.Correct*a.Total
	})
	slices.Reverse(out.Bands)
	return out, nil
}
