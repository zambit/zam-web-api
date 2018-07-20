package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gitlab.com/ZamzamTech/wallet-api/services/sessions"
	"net/http"
	"strings"
)

// AuthMiddlewareFactory creates auth middleware using session validation via given storage
func AuthMiddlewareFactory(
	sessStorage sessions.IStorage,
	tokenName string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		authToken, err := GetAuthTokenFromContext(c, tokenName)
		if err != nil {
			abortUnauthorized(c, err.Error())
			return
		}

		// ensure storage have this token
		data, err := sessStorage.Get(sessions.Token(authToken))
		if err != nil {
			switch err {
			case sessions.ErrNotFound, sessions.ErrUnexpectedToken, sessions.ErrExpired:
				abortUnauthorized(c, err.Error())
			default:
				abortMiddlware(c, http.StatusInternalServerError, "token validation failed")
			}
			return
		}

		// attach user data to context
		c.Set("user_data", data)

		// continue handlers
		c.Next()
	}
}

// GetAuthTokenFromContext gets auth token from request headers or return error.
// Temporary placed here.
func GetAuthTokenFromContext(c *gin.Context, tokenName string) (string, error) {
	// get token
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("Authorization header is empty")
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != tokenName {
		return "", fmt.Errorf("auth header is wrong, expect %s token", tokenName)
	}
	return parts[1], nil
}

// GetUserDataFromContext gets user data extracted from session storage by auth-token during middleware work
func GetUserDataFromContext(c *gin.Context) map[string]interface{} {
	return c.GetStringMap("user_data")
}

//
func abortUnauthorized(c *gin.Context, message string) {
	abortMiddlware(c, http.StatusUnauthorized, message)
}

// setUnauthorized
func abortMiddlware(c *gin.Context, code int, message string) {
	c.JSON(code, map[string]interface{}{
		"result": false,
		"errors": []interface{}{
			map[string]interface{}{
				"message": message,
			},
		},
	})
	c.Abort()
}
