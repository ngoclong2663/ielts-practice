package usecases

import (
	"context"

	"ielts-practice/app/modules/feedback/domain/entity"
	"ielts-practice/app/modules/feedback/dto"
)

// ListFeedback is the admin feedback list: newest first; counts ignore the status filter.
func (uc *FeedbackUseCases) ListFeedback(ctx context.Context, q dto.ListFeedbackQueryDto) (dto.FeedbackListDto, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = defaultPageSize
	}
	f := entity.FeedbackFilter{Status: q.Status, Category: q.Category, Page: q.Page, PageSize: q.PageSize}
	if f.Status == "" {
		f.Status = entity.StatusNew
	}
	if f.Status == "all" {
		f.Status = ""
	}
	if f.Category == "all" {
		f.Category = ""
	}

	items, total, err := uc.repo.List(ctx, f)
	if err != nil {
		return dto.FeedbackListDto{}, internalErr("Failed to list feedback", err)
	}
	counts, err := uc.repo.CountByStatus(ctx, f.Category)
	if err != nil {
		return dto.FeedbackListDto{}, internalErr("Failed to count feedback", err)
	}
	out := dto.FeedbackListDto{
		Items: make([]dto.FeedbackItemDto, 0, len(items)), Total: total, Page: q.Page, PageSize: q.PageSize,
		Counts: dto.StatusCountsDto{New: counts.New, Seen: counts.Seen, Resolved: counts.Resolved},
	}
	for i := range items {
		out.Items = append(out.Items, toItemDto(&items[i]))
	}
	return out, nil
}
