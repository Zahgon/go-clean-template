package v1

import (
	"github.com/evrone/go-clean-template/internal/controller/restapi/middleware"
	"github.com/gin-gonic/gin"
)

// extractUserID returns the user ID stored in the Gin context by the auth middleware.
func extractUserID(ctx *gin.Context) (string, bool) {
	value, exists := ctx.Get(middleware.UserIDKey)
	if !exists {
		return "", false
	}

	userID, ok := value.(string)

	return userID, ok
}
