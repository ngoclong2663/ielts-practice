package internal

import (
	"github.com/gin-gonic/gin"
	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/auth/dto"
	"ielts-practice/configs"
)

type AuthRouter struct {
	presenter *AuthPresenter
	cfg       *configs.Config
}

func RegisterRoutes(router *gin.Engine, ar *AuthRouter) {
	ar.Register(router)
}

func NewAuthRouter(presenter *AuthPresenter, cfg *configs.Config) *AuthRouter {
	return &AuthRouter{presenter: presenter, cfg: cfg}
}

func (ar *AuthRouter) Register(router *gin.Engine) {
	authGroup := router.Group("/api/v1/auth", middlewares.MediumTermRateLimitMiddleware())
	{
		authGroup.POST("/google", middlewares.ValidateRequestBody[dto.GoogleSignInRequestDto](), ar.presenter.SignInWithGoogle)
		authGroup.POST("/refresh", middlewares.ValidateRequestBody[dto.RefreshRequestDto](), ar.presenter.Refresh)
		authGroup.POST("/logout", middlewares.ValidateRequestBody[dto.LogoutRequestDto](), ar.presenter.Logout)

		if ar.cfg.IsTest() {
			authGroup.POST("/test-login", middlewares.ValidateRequestBody[dto.TestLoginRequestDto](), ar.presenter.TestLogin)
		}
	}
}
