package middleware

import (
	"net/http"
	"strings"

	"github.com/evrone/go-clean-template/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const _bearerParts = 2

// UserIDKey is the Gin context key holding the authenticated user ID.
const UserIDKey = "userID"

type errorResponse struct {
	Error string `json:"error"`
}

// Auth returns a JWT authentication middleware for Gin.
func Auth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if header == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{Error: "missing authorization header"})

			return
		}

		parts := strings.SplitN(header, " ", _bearerParts)
		if len(parts) != _bearerParts || parts[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{Error: "invalid authorization header format"})

			return
		}

		userID, err := jwtManager.ParseToken(parts[1])
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{Error: "invalid or expired token"})

			return
		}

		ctx.Set(UserIDKey, userID)

		ctx.Next()
	}
}
