package usecases

import (
	"context"
	"strings"

	"ielts-practice/app/modules/admin/domain/entity"
	"ielts-practice/app/modules/admin/dto"
)

// ListUsers: every user with their activity and AI use; search by email or name.
func (uc *AdminUseCases) ListUsers(ctx context.Context, q dto.ListUsersQueryDto) (dto.AdminUserListDto, error) {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = defaultPageSize
	}
	if q.Sort == "" {
		q.Sort = entity.SortLastActive
	}
	users, total, err := uc.repo.ListUsers(ctx, entity.UserFilter{
		Query: strings.TrimSpace(q.Q), Sort: q.Sort, Page: q.Page, PageSize: q.PageSize, Since7d: uc.window().Since7d,
	})
	if err != nil {
		return dto.AdminUserListDto{}, internalErr("Failed to list users", err)
	}
	out := dto.AdminUserListDto{Items: make([]dto.AdminUserDto, 0, len(users)), Total: total, Page: q.Page, PageSize: q.PageSize}
	for _, u := range users {
		out.Items = append(out.Items, dto.AdminUserDto{
			ID: u.ID, Email: u.Email, DisplayName: u.DisplayName, AvatarURL: u.AvatarURL, UILocale: u.UILocale,
			CreatedAt: u.CreatedAt, LastLoginAt: u.LastLoginAt, LastActiveAt: u.LastActiveAt,
			Decks: u.Decks, Cards: u.Cards, ReviewsTotal: u.ReviewsTotal, Reviews7d: u.Reviews7d,
			AIWords: u.AIWords, TopicLists: u.TopicLists,
			WritingAttempts: u.WritingAttempts, WritingSubmitted: u.WritingSubmit, WritingAICalls: u.WritingAICalls,
			Feedback: u.Feedback,
		})
	}
	return out, nil
}
