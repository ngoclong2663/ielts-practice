package mapper

import (
	"encoding/json"

	"ielts-practice/app/models"
	"ielts-practice/app/modules/vocab/domain/entity"
	"ielts-practice/app/modules/vocab/dto"
)

type VocabMapper struct{}

func NewVocabMapper() *VocabMapper { return &VocabMapper{} }

func MarshalJSON(v any, empty string) []byte {
	b, err := json.Marshal(v)
	if err != nil || string(b) == "null" {
		return []byte(empty)
	}
	return b
}

func unmarshalContent(b []byte) *entity.CardContent {
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	var c entity.CardContent
	if err := json.Unmarshal(b, &c); err != nil {
		return nil
	}
	return &c
}

func unmarshalImages(b []byte) []entity.CardImage {
	out := []entity.CardImage{}
	if len(b) > 0 {
		_ = json.Unmarshal(b, &out)
	}
	return out
}

func (m *VocabMapper) LexemeModelToEntity(l *models.Lexeme) *entity.Lexeme {
	flags := []string{}
	if len(l.QualityFlags) > 0 {
		_ = json.Unmarshal(l.QualityFlags, &flags)
	}
	return &entity.Lexeme{
		ID:              l.ID,
		Key:             l.Key,
		Headword:        l.Headword,
		Lemma:           l.Lemma,
		SenseHint:       l.SenseHint,
		MeaningLang:     l.MeaningLang,
		Content:         unmarshalContent(l.Content),
		Status:          l.Status,
		Suggestion:      l.Suggestion,
		ImageQuery:      l.ImageQuery,
		ImageCandidates: unmarshalImages(l.ImageCandidates),
		QualityFlags:    flags,
		ReportCount:     l.ReportCount,
	}
}

func (m *VocabMapper) JobModelToEntity(j *models.GenerationJob) *entity.Job {
	return &entity.Job{
		ID:              j.ID,
		UserID:          j.UserID,
		TargetDeckID:    j.TargetDeckID,
		Status:          j.Status,
		Total:           j.Total,
		Done:            j.Done,
		CommittedDeckID: j.CommittedDeckID,
		CreatedAt:       j.CreatedAt,
		FinishedAt:      j.FinishedAt,
	}
}

func (m *VocabMapper) ItemModelToEntity(i *models.GenerationJobItem) *entity.JobItem {
	return &entity.JobItem{
		ID:              i.ID,
		JobID:           i.JobID,
		UserID:          i.UserID,
		Position:        i.Position,
		InputText:       i.InputText,
		SenseHint:       i.SenseHint,
		LexemeID:        i.LexemeID,
		Status:          i.Status,
		Regenerate:      i.Regenerate,
		Charged:         i.Charged,
		Cached:          i.Cached,
		Content:         unmarshalContent(i.Content),
		ImageCandidates: unmarshalImages(i.ImageCandidates),
		ErrorCode:       i.ErrorCode,
		Suggestion:      i.Suggestion,
		Attempts:        i.Attempts,
		NextAttemptAt:   i.NextAttemptAt,
		Removed:         i.Removed,
		UpdatedAt:       i.UpdatedAt,
	}
}

func (m *VocabMapper) ImageLinkModelToEntity(l *models.ImageLink) *entity.ImageLink {
	var replacement *entity.CardImage
	if len(l.Replacement) > 0 && string(l.Replacement) != "null" {
		var img entity.CardImage
		if json.Unmarshal(l.Replacement, &img) == nil {
			replacement = &img
		}
	}
	return &entity.ImageLink{
		Provider:    l.Provider,
		ProviderID:  l.ProviderID,
		URL:         l.URL,
		ThumbURL:    l.ThumbURL,
		ExpiresAt:   l.ExpiresAt,
		Query:       l.Query,
		Headword:    l.Headword,
		Replacement: replacement,
		Gone:        l.Gone,
	}
}

func (m *VocabMapper) ItemEntityToDto(i *entity.JobItem) dto.JobItemDto {
	candidates := i.ImageCandidates
	if candidates == nil {
		candidates = []entity.CardImage{}
	}
	return dto.JobItemDto{
		ID:              i.ID,
		Position:        i.Position,
		InputText:       i.InputText,
		SenseHint:       i.SenseHint,
		Status:          i.Status,
		ErrorCode:       i.ErrorCode,
		Suggestion:      i.Suggestion,
		Cached:          i.Cached,
		LexemeID:        i.LexemeID,
		Content:         i.Content,
		ImageCandidates: candidates,
		Removed:         i.Removed,
	}
}

func (m *VocabMapper) JobEntityToDto(j *entity.Job) dto.JobDto {
	items := make([]dto.JobItemDto, 0, len(j.Items))
	for i := range j.Items {
		items = append(items, m.ItemEntityToDto(&j.Items[i]))
	}
	return dto.JobDto{
		ID:              j.ID,
		Status:          j.Status,
		Total:           j.Total,
		Done:            j.Done,
		TargetDeckID:    j.TargetDeckID,
		CommittedDeckID: j.CommittedDeckID,
		CreatedAt:       j.CreatedAt,
		FinishedAt:      j.FinishedAt,
		Items:           items,
	}
}

func (m *VocabMapper) JobEntityToSummaryDto(j *entity.Job) dto.JobSummaryDto {
	words := []string{}
	for _, it := range j.Items {
		if !it.Removed {
			words = append(words, it.InputText)
		}
	}
	return dto.JobSummaryDto{
		ID:              j.ID,
		Status:          j.Status,
		Total:           j.Total,
		Done:            j.Done,
		CommittedDeckID: j.CommittedDeckID,
		CreatedAt:       j.CreatedAt,
		FinishedAt:      j.FinishedAt,
		Words:           words,
	}
}

func (m *VocabMapper) SuggestionModelToEntity(s *models.TopicSuggestion) *entity.Suggestion {
	words := []entity.SuggestedWord{}
	if len(s.Words) > 0 {
		_ = json.Unmarshal(s.Words, &words)
	}
	return &entity.Suggestion{
		ID: s.ID, UserID: s.UserID, DeckID: s.DeckID, TopicID: s.TopicID, TopicText: s.TopicText,
		Requested: s.Requested, Words: words, SkippedExisting: s.SkippedExisting, CreatedAt: s.CreatedAt,
	}
}
