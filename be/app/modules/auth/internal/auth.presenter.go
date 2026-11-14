package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ielts-practice/app/modules/auth/domain/entity"
	"ielts-practice/app/modules/auth/dto"
	usecases "ielts-practice/app/modules/auth/internal/use-cases"
	"ielts-practice/pkg"
)

type AuthPresenter struct {
	uc *usecases.AuthUseCases
}

func NewAuthPresenter(uc *usecases.AuthUseCases) *AuthPresenter {
	return &AuthPresenter{uc: uc}
}

func clientMeta(c *gin.Context) entity.ClientMeta {
	return entity.ClientMeta{UserAgent: c.Request.UserAgent(), IP: c.ClientIP()}
}

// SignInWithGoogle: POST /api/v1/auth/google
func (p *AuthPresenter) SignInWithGoogle(c *gin.Context) {
	req := pkg.MustGetBody[dto.GoogleSignInRequestDto](c)
	session, err := p.uc.SignInWithGoogle(c.Request.Context(), req, clientMeta(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, session)
}

// Refresh: POST /api/v1/auth/refresh
func (p *AuthPresenter) Refresh(c *gin.Context) {
	req := pkg.MustGetBody[dto.RefreshRequestDto](c)
	session, err := p.uc.Refresh(c.Request.Context(), req.RefreshToken, clientMeta(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, session)
}

// Logout: POST /api/v1/auth/logout
func (p *AuthPresenter) Logout(c *gin.Context) {
	req := pkg.MustGetBody[dto.LogoutRequestDto](c)
	if err := p.uc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// TestLogin: POST /api/v1/auth/test-login (APP_ENV=test only)
func (p *AuthPresenter) TestLogin(c *gin.Context) {
	req := pkg.MustGetBody[dto.TestLoginRequestDto](c)
	session, err := p.uc.TestLogin(c.Request.Context(), req, clientMeta(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, session)
}
