package internal

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine, p *WritingPresenter) {
	// Public: figures are task content (no user data), and <img> tags can't refresh an expired session.
	router.GET("/api/v1/writing/figures/:file", p.ResolveFigure)
}
