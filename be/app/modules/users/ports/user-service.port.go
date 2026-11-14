package ports

import (
	"context"

	"ielts-practice/app/modules/users/domain/entity"
)

// GoogleUserInput is the verified identity coming from Google plus client hints for first sign-in.
type GoogleUserInput struct {
	GoogleSub   string
	Email       string
	DisplayName string
	AvatarURL   string
	UILocale    string // used only when the user is created
	Timezone    string // used only when the user is created
}

type IUserService interface {
	// UpsertGoogleUser creates the user (with default study settings) on first sign-in,
	// otherwise refreshes email, name, avatar and last_login_at. Call it inside a transaction.
	UpsertGoogleUser(ctx context.Context, input GoogleUserInput) (*entity.UserEntity, error)
	GetUserByID(ctx context.Context, id string) (*entity.UserEntity, error)
}
