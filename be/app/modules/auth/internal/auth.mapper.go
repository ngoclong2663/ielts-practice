package internal

import (
	"ielts-practice/app/modules/auth/domain/entity"
	"ielts-practice/app/modules/auth/dto"
	userEntity "ielts-practice/app/modules/users/domain/entity"
	"ielts-practice/app/modules/users/ports"
)

type AuthMapper struct{}

func NewAuthMapper() *AuthMapper {
	return &AuthMapper{}
}

func (m *AuthMapper) MapGoogleIdentityToUserInput(id *entity.GoogleIdentity, req dto.GoogleSignInRequestDto) ports.GoogleUserInput {
	return ports.GoogleUserInput{
		GoogleSub:   id.Subject,
		Email:       id.Email,
		DisplayName: id.Name,
		AvatarURL:   id.Picture,
		UILocale:    req.UILocale,
		Timezone:    req.Timezone,
	}
}

func (m *AuthMapper) MapUserToSessionUser(u *userEntity.UserEntity) dto.SessionUserDto {
	return dto.SessionUserDto{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		UILocale:    u.UILocale,
	}
}
