package mapper

import (
	"encoding/json"

	"ielts-practice/app/models"
	"ielts-practice/app/modules/decks/domain/entity"
	"ielts-practice/app/modules/decks/dto"
)

type DeckMapper struct{}

func NewDeckMapper() *DeckMapper {
	return &DeckMapper{}
}

func (m *DeckMapper) DeckModelToEntity(d *models.Deck, cardCount int) *entity.DeckEntity {
	return &entity.DeckEntity{
		ID:            d.ID,
		UserID:        d.UserID,
		Name:          d.Name,
		Description:   d.Description,
		Source:        d.Source,
		CollectionID:  d.CollectionID,
		NewPerDay:     d.NewPerDay,
		ReviewsPerDay: d.ReviewsPerDay,
		CardCount:     cardCount,
		CreatedAt:     d.CreatedAt,
	}
}

func (m *DeckMapper) DeckEntityToModel(e *entity.DeckEntity) *models.Deck {
	return &models.Deck{
		ID:            e.ID,
		UserID:        e.UserID,
		Name:          e.Name,
		Description:   e.Description,
		Source:        e.Source,
		CollectionID:  e.CollectionID,
		NewPerDay:     e.NewPerDay,
		ReviewsPerDay: e.ReviewsPerDay,
	}
}

func (m *DeckMapper) DeckEntityToDto(e *entity.DeckEntity) dto.DeckDto {
	return dto.DeckDto{
		ID:            e.ID,
		Name:          e.Name,
		Description:   e.Description,
		Source:        e.Source,
		CollectionID:  e.CollectionID,
		NewPerDay:     e.NewPerDay,
		ReviewsPerDay: e.ReviewsPerDay,
		CardCount:     e.CardCount,
		CreatedAt:     e.CreatedAt,
	}
}

func (m *DeckMapper) CardModelToEntity(c *models.Card) (*entity.CardEntity, error) {
	var content entity.CardContent
	if err := json.Unmarshal(c.Content, &content); err != nil {
		return nil, err
	}
	return &entity.CardEntity{
		ID:            c.ID,
		UserID:        c.UserID,
		DeckID:        c.DeckID,
		LexemeID:      c.LexemeID,
		Headword:      c.Headword,
		Content:       content,
		State:         c.State,
		Due:           c.Due,
		Position:      c.Position,
		ScheduledDays: c.ScheduledDays,
		Reps:          c.Reps,
		Lapses:        c.Lapses,
		LastReviewAt:  c.LastReviewAt,
		Suspended:     c.Suspended,
		IsLeech:       c.IsLeech,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}, nil
}

func (m *DeckMapper) CardEntityToSummaryDto(e *entity.CardEntity) dto.CardSummaryDto {
	return dto.CardSummaryDto{
		ID:            e.ID,
		DeckID:        e.DeckID,
		Headword:      e.Headword,
		MeaningVI:     e.Content.MeaningVI,
		State:         entity.StateName(e.State),
		Due:           e.Due,
		ScheduledDays: e.ScheduledDays,
		Reps:          e.Reps,
		Lapses:        e.Lapses,
		Suspended:     e.Suspended,
		IsLeech:       e.IsLeech,
		CreatedAt:     e.CreatedAt,
	}
}

func (m *DeckMapper) CardEntityToDto(e *entity.CardEntity) dto.CardDto {
	return dto.CardDto{CardSummaryDto: m.CardEntityToSummaryDto(e), Content: e.Content}
}

// ApplyManualFields merges the manual-card form into existing content, keeping
// generated fields (word forms, collocations, audio) untouched.
func (m *DeckMapper) ApplyManualFields(content *entity.CardContent, f dto.ManualCardFields) {
	if f.Headword != nil {
		content.Headword = *f.Headword
	}
	if f.MeaningVI != nil {
		content.MeaningVI = *f.MeaningVI
	}
	if f.PartOfSpeech != nil {
		content.PartOfSpeech = *f.PartOfSpeech
	}
	if f.IPA != nil {
		if *f.IPA == "" {
			content.IPA = nil
		} else {
			if content.IPA == nil {
				content.IPA = &entity.IPA{}
			}
			content.IPA.UK = *f.IPA
		}
	}
	if f.ExampleEN != nil || f.ExampleVI != nil {
		var ex entity.Example
		if len(content.ExtraExamples) > 0 {
			ex = content.ExtraExamples[0]
		}
		if f.ExampleEN != nil {
			ex.EN = *f.ExampleEN
		}
		if f.ExampleVI != nil {
			ex.VI = *f.ExampleVI
		}
		switch {
		case ex.EN == "" && len(content.ExtraExamples) > 0:
			content.ExtraExamples = content.ExtraExamples[1:]
		case ex.EN == "":
		case len(content.ExtraExamples) > 0:
			content.ExtraExamples[0] = ex
		default:
			content.ExtraExamples = []entity.Example{ex}
		}
	}
	if f.ImageURL != nil {
		if *f.ImageURL == "" {
			content.Image = nil
		} else {
			content.Image = &entity.CardImage{Provider: "manual", URL: *f.ImageURL, Alt: content.Headword}
		}
	}
}
