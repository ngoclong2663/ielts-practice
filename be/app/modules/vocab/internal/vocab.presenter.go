package internal

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"ielts-practice/app/middlewares"
	"ielts-practice/app/modules/vocab/dto"
	usecases "ielts-practice/app/modules/vocab/internal/use-cases"
	"ielts-practice/pkg"
	"ielts-practice/pkg/exceptions"
)

type VocabPresenter struct {
	uc *usecases.VocabUseCases
}

func NewVocabPresenter(uc *usecases.VocabUseCases) *VocabPresenter {
	return &VocabPresenter{uc: uc}
}

func respond[T any](c *gin.Context, status int, data T, err error) {
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(status, data)
}

func (p *VocabPresenter) GetStatus(c *gin.Context) {
	data, err := p.uc.GetStatus(c.Request.Context(), middlewares.MustGetUserID(c))
	respond(c, http.StatusOK, data, err)
}

func (p *VocabPresenter) Normalize(c *gin.Context) {
	c.JSON(http.StatusOK, p.uc.Normalize(pkg.MustGetBody[dto.NormalizeRequestDto](c)))
}

func (p *VocabPresenter) CreateJob(c *gin.Context) {
	data, err := p.uc.CreateJob(c.Request.Context(), middlewares.MustGetUserID(c), pkg.MustGetBody[dto.CreateJobRequestDto](c))
	respond(c, http.StatusCreated, data, err)
}

func (p *VocabPresenter) ListJobs(c *gin.Context) {
	data, err := p.uc.ListJobs(c.Request.Context(), middlewares.MustGetUserID(c))
	respond(c, http.StatusOK, data, err)
}

func (p *VocabPresenter) GetJob(c *gin.Context) {
	data, err := p.uc.GetJob(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("jobId"))
	respond(c, http.StatusOK, data, err)
}

func (p *VocabPresenter) RetryItem(c *gin.Context) {
	data, err := p.uc.RetryItem(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("jobId"), c.Param("itemId"),
		pkg.MustGetBody[dto.RetryItemRequestDto](c))
	respond(c, http.StatusOK, data, err)
}

func (p *VocabPresenter) SearchImages(c *gin.Context) {
	q, _ := pkg.GetQuery[dto.ImageSearchQueryDto](c)
	data, err := p.uc.SearchImages(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("jobId"), c.Param("itemId"), q)
	respond(c, http.StatusOK, data, err)
}

func (p *VocabPresenter) UpdateItem(c *gin.Context) {
	data, err := p.uc.UpdateItem(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("jobId"), c.Param("itemId"),
		pkg.MustGetBody[dto.UpdateItemRequestDto](c))
	respond(c, http.StatusOK, data, err)
}

func (p *VocabPresenter) CommitJob(c *gin.Context) {
	data, err := p.uc.CommitJob(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("jobId"), pkg.MustGetBody[dto.CommitRequestDto](c))
	respond(c, http.StatusCreated, data, err)
}

func (p *VocabPresenter) ReportLexeme(c *gin.Context) {
	if err := p.uc.ReportLexeme(c.Request.Context(), middlewares.MustGetUserID(c), c.Param("lexemeId"), pkg.MustGetBody[dto.ReportRequestDto](c)); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

var (
	providerParam = regexp.MustCompile(`^[a-z]{1,16}$`)
	idParam       = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
)

// ResolveImage redirects to the provider's current direct link (Epic 3 §5.4).
func (p *VocabPresenter) ResolveImage(c *gin.Context) {
	provider, id := c.Param("provider"), c.Param("id")
	if !providerParam.MatchString(provider) || !idParam.MatchString(id) {
		_ = c.Error(exceptions.NewException(http.StatusNotFound, exceptions.CodeImageNotFound, "Image not found", nil))
		return
	}
	redirect, err := p.uc.ResolveImage(c.Request.Context(), provider, id, c.Query("size") == "thumb")
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", int(redirect.MaxAge.Seconds())))
	c.Redirect(http.StatusFound, redirect.Location)
}
