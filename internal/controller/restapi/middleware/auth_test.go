package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evrone/go-clean-template/internal/controller/restapi/middleware"
	"github.com/evrone/go-clean-template/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestApp(t *testing.T) (*gin.Engine, *jwt.Manager) {
	t.Helper()

	jwtManager := jwt.New("test-secret", time.Hour)

	gin.SetMode(gin.TestMode)

	app := gin.New()
	app.Use(middleware.Auth(jwtManager))
	app.GET("/test", func(c *gin.Context) {
		userID, ok := c.Get(middleware.UserIDKey)
		if !ok {
			c.Status(http.StatusUnauthorized)

			return
		}

		c.String(http.StatusOK, "%s", userID)
	})

	return app, jwtManager
}

func TestAuthMiddleware(t *testing.T) {
	t.Parallel()

	app, jwtManager := newTestApp(t)

	validToken, err := jwtManager.GenerateToken("user-id-123")
	require.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "missing header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid format",
			authHeader:     "Basic xxx",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectedBody:   "user-id-123",
		},
	}

	for _, tc := range tests {
		localTc := tc

		t.Run(localTc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/test", http.NoBody)
			if localTc.authHeader != "" {
				req.Header.Set("Authorization", localTc.authHeader)
			}

			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			resp := rec.Result()

			defer resp.Body.Close()

			assert.Equal(t, localTc.expectedStatus, resp.StatusCode)

			if localTc.expectedBody != "" {
				body, readErr := io.ReadAll(resp.Body)
				require.NoError(t, readErr)
				assert.Equal(t, localTc.expectedBody, string(body))
			}
		})
	}
}
