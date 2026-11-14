package internal

import (
	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/users/dto"
	"ielts-practice/pkg/tokens"

	"github.com/gin-gonic/gin"
)

type UserRouter struct {
	presenter *UserPresenter
	jwt       *tokens.JWTManager
}

func RegisterRoutes(router *gin.Engine, ur *UserRouter) {
	ur.Register(router)
}

func NewUserRouter(presenter *UserPresenter, jwt *tokens.JWTManager) *UserRouter {
	return &UserRouter{presenter: presenter, jwt: jwt}
}

func (ur *UserRouter) Register(router *gin.Engine) {
	me := router.Group("/api/v1/me", middlewares.AuthMiddleware(ur.jwt))
	{
		me.GET("", ur.presenter.GetMe)
		me.PATCH("", middlewares.ValidateRequestBody[dto.UpdateProfileRequestDto](), ur.presenter.UpdateProfile)
		me.PATCH("/study-settings", middlewares.ValidateRequestBody[dto.UpdateStudySettingsRequestDto](), ur.presenter.UpdateStudySettings)
	}
}
