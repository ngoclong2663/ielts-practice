package mapper

import (
	"ielts-practice/app/modules/auth/domain/entity"
	"ielts-practice/app/modules/auth/dto"
	userEntity "ielts-practice/app/modules/users/domain/entity"
	usersin "ielts-practice/app/modules/users/ports/in"
)

type AuthMapper struct{}

func NewAuthMapper() *AuthMapper {
	return &AuthMapper{}
}

func (m *AuthMapper) MapGoogleIdentityToUserInput(id *entity.GoogleIdentity, req dto.GoogleSignInRequestDto) usersin.GoogleUserInput {
	return usersin.GoogleUserInput{
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
