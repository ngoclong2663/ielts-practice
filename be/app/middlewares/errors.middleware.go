package middlewares

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"ielts-practice/pkg/exceptions"
)

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Params  map[string]any    `json:"params,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// WriteError writes the standard error envelope: { "error": { code, message, params, fields } }.
func WriteError(c *gin.Context, appErr *exceptions.AppError) {
	c.AbortWithStatusJSON(appErr.Status, gin.H{"error": errorBody{
		Code:    appErr.Code,
		Message: appErr.Message,
		Params:  appErr.Params,
		Fields:  appErr.Fields,
	}})
}

func ErrorMiddleware(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		err := c.Errors.Last().Err

		var appErr *exceptions.AppError
		if errors.As(err, &appErr) {
			if appErr.Status >= http.StatusInternalServerError {
				logger.Error().Err(appErr.Err).Str("request_id", GetRequestID(c)).Str("code", appErr.Code).Msg(appErr.Message)
			}
			WriteError(c, appErr)
			return
		}

		logger.Error().Err(err).Str("request_id", GetRequestID(c)).Msg("unhandled error")
		WriteError(c, exceptions.ErrInternalServerError("Internal server error", err))
	}
}
