package internal

import (
	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/admin/dto"
	adminin "ielts-practice/app/modules/admin/ports/in"
)

type AdminRouter struct {
	presenter *AdminPresenter
	session   adminin.IAdminSession
}

func RegisterRoutes(router *gin.Engine, r *AdminRouter) {
	r.Register(router)
}

func NewAdminRouter(presenter *AdminPresenter, session adminin.IAdminSession) *AdminRouter {
	return &AdminRouter{presenter: presenter, session: session}
}

func (r *AdminRouter) Register(router *gin.Engine) {
	router.POST("/api/v1/admin/auth/login", middlewares.MediumTermRateLimitMiddleware(),
		middlewares.ValidateRequestBody[dto.LoginRequestDto](), r.presenter.Login)

	admin := router.Group("/api/v1/admin", middlewares.AdminGuard(r.session)...)
	admin.GET("/auth/me", r.presenter.GetMe)
	admin.GET("/overview", r.presenter.GetOverview)
	admin.GET("/users", middlewares.ValidateQueryParams[dto.ListUsersQueryDto](), r.presenter.ListUsers)
}
