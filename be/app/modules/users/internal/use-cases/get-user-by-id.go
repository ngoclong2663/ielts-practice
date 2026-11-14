package usecases

import (
	"context"

	"ielts-practice/app/modules/users/domain/entity"
	"ielts-practice/pkg/exceptions"
)

func (u *UserUseCases) GetUserByID(ctx context.Context, id string) (*entity.UserEntity, error) {
	user, err := u.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, exceptions.ErrInternalServerError("Failed to load user", err)
	}
	if user == nil {
		return nil, exceptions.ErrNotFound("User not found", nil)
	}
	return user, nil
}
