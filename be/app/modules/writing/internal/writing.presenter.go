package internal

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	usecases "ielts-practice/app/modules/writing/internal/use-cases"
)

type WritingPresenter struct {
	uc *usecases.WritingUseCases
}

func NewWritingPresenter(uc *usecases.WritingUseCases) *WritingPresenter {
	return &WritingPresenter{uc: uc}
}

// ResolveFigure redirects to a presigned S3 link for a task figure.
func (p *WritingPresenter) ResolveFigure(c *gin.Context) {
	url, err := p.uc.ResolveFigure(c.Request.Context(), c.Param("file"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", int(usecases.FigureCacheMaxAge.Seconds())))
	c.Redirect(http.StatusFound, url)
}
