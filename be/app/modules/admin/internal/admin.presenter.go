package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/admin/dto"
	usecases "ielts-practice/app/modules/admin/internal/use-cases"
	"ielts-practice/pkg"
)

type AdminPresenter struct {
	uc usecases.IAdminUseCases
}

func NewAdminPresenter(uc usecases.IAdminUseCases) *AdminPresenter {
	return &AdminPresenter{uc: uc}
}

func respond[T any](c *gin.Context, data T, err error) {
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, data)
}

func (p *AdminPresenter) GetOverview(c *gin.Context) {
	data, err := p.uc.GetOverview(c.Request.Context())
	respond(c, data, err)
}

func (p *AdminPresenter) ListUsers(c *gin.Context) {
	q, _ := pkg.GetQuery[dto.ListUsersQueryDto](c)
	data, err := p.uc.ListUsers(c.Request.Context(), q)
	respond(c, data, err)
}

func (p *AdminPresenter) Login(c *gin.Context) {
	req := pkg.MustGetBody[dto.LoginRequestDto](c)
	data, err := p.uc.Login(c.Request.Context(), req)
	respond(c, data, err)
}

func (p *AdminPresenter) GetMe(c *gin.Context) {
	data, err := p.uc.GetMe(c.Request.Context(), middlewares.MustGetAdminID(c))
	respond(c, data, err)
}
