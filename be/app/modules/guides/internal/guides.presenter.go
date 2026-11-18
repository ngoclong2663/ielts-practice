package internal

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"ielts-practice/app/modules/guides/dto"
	usecases "ielts-practice/app/modules/guides/internal/use-cases"
	"ielts-practice/pkg"
)

type GuidesPresenter struct {
	uc *usecases.GuidesUseCases
}

func NewGuidesPresenter(uc *usecases.GuidesUseCases) *GuidesPresenter {
	return &GuidesPresenter{uc: uc}
}

func (p *GuidesPresenter) ListGuides(c *gin.Context) {
	q, _ := pkg.GetQuery[dto.ListGuidesQueryDto](c)
	data, err := p.uc.ListGuides(c.Request.Context(), q.Topic)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, data)
}

// ResolveMedia redirects to a presigned S3 link for a guide's video or poster.
func (p *GuidesPresenter) ResolveMedia(c *gin.Context) {
	url, err := p.uc.ResolveMedia(c.Request.Context(), c.Param("file"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", int(usecases.MediaCacheMaxAge.Seconds())))
	c.Redirect(http.StatusFound, url)
}
