package internal

import (
	"net/http"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/users/dto"
	usecases "ielts-practice/app/modules/users/internal/use-cases"
	"ielts-practice/pkg"

	"github.com/gin-gonic/gin"
)

type UserPresenter struct {
	uc usecases.IUserUseCases
}

func NewUserPresenter(uc usecases.IUserUseCases) *UserPresenter {
	return &UserPresenter{uc: uc}
}

func (p *UserPresenter) GetMe(c *gin.Context) {
	data, err := p.uc.GetMe(c.Request.Context(), middlewares.MustGetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func (p *UserPresenter) UpdateProfile(c *gin.Context) {
	req := pkg.MustGetBody[dto.UpdateProfileRequestDto](c)
	data, err := p.uc.UpdateProfile(c.Request.Context(), middlewares.MustGetUserID(c), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func (p *UserPresenter) UpdateStudySettings(c *gin.Context) {
	req := pkg.MustGetBody[dto.UpdateStudySettingsRequestDto](c)
	data, err := p.uc.UpdateStudySettings(c.Request.Context(), middlewares.MustGetUserID(c), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, data)
}
