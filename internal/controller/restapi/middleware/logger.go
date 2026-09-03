package middleware

import (
	"strconv"
	"strings"

	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gin-gonic/gin"
)

func buildRequestMessage(ctx *gin.Context) string {
	var result strings.Builder

	size := max(ctx.Writer.Size(), 0)

	result.WriteString(ctx.ClientIP())
	result.WriteString(" - ")
	result.WriteString(ctx.Request.Method)
	result.WriteString(" ")
	result.WriteString(ctx.Request.URL.RequestURI())
	result.WriteString(" - ")
	result.WriteString(strconv.Itoa(ctx.Writer.Status()))
	result.WriteString(" ")
	result.WriteString(strconv.Itoa(size))

	return result.String()
}

func Logger(l logger.Interface) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		l.Info("%s", buildRequestMessage(ctx))
	}
}
