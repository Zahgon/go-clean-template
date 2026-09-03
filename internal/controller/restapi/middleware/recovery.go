package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/gin-gonic/gin"
)

func buildPanicMessage(ctx *gin.Context, err any) string {
	var result strings.Builder

	result.WriteString(ctx.ClientIP())
	result.WriteString(" - ")
	result.WriteString(ctx.Request.Method)
	result.WriteString(" ")
	result.WriteString(ctx.Request.URL.RequestURI())
	result.WriteString(" PANIC DETECTED: ")
	fmt.Fprintf(&result, "%v\n%s\n", err, debug.Stack())

	return result.String()
}

func logPanic(l logger.Interface) gin.RecoveryFunc {
	return func(ctx *gin.Context, err any) {
		l.Error(buildPanicMessage(ctx, err))

		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
}

func Recovery(l logger.Interface) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(nil, logPanic(l))
}
