package dto

import "ielts-practice/app/modules/guides/domain/entity"

type ListGuidesQueryDto struct {
	Topic string `form:"topic" binding:"omitempty,oneof=vocab writing"`
}

type GuideDto struct {
	Slug            string           `json:"slug"`
	Topic           string           `json:"topic"`
	Lang            string           `json:"lang"` // language of the narration
	Title           entity.Localized `json:"title"`
	Description     entity.Localized `json:"description"`
	DurationSeconds float64          `json:"duration_seconds"`
	VideoURL        string           `json:"video_url"`
	PosterURL       string           `json:"poster_url"` // empty when the guide has no poster
}
