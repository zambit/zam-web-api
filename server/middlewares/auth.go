package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin"
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
		// get token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			abortUnauthorized(c, "auth header is empty")
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] == tokenName {
			abortUnauthorized(c, fmt.Sprintf("auth header is wrong, expect %s token", tokenName))
		}

		// ensure storage have this token
		data, err := sessStorage.Get(sessions.Token(parts[1]))
		if err != nil {
			switch err {
			case sessions.ErrNotFound, sessions.ErrUnexpectedToken, sessions.ErrExpired:
				abortUnauthorized(c, err.Error())
			default:
				abortMiddlware(c, http.StatusInternalServerError, "token validation failed")
			}
		}

		// attach user data to context
		c.Set("user_data", data)

		// continue handlers
		c.Next()
	}
}

// GetUserDataFromContext gets user data extracted from session storage by auth-token during middleware work
func GetUserDataFromContext(c *gin.Context) interface{} {
	val, exists := c.Get("user_data")
	if !exists {
		return nil
	}
	return val
}

//
func abortUnauthorized(c *gin.Context, message string) {
	abortMiddlware(c, http.StatusUnauthorized, message)
}

// setUnauthorized
func abortMiddlware(c *gin.Context, code int, message string) {
	c.JSON(code, map[string]interface{}{
		"code":    code,
		"message": message,
	})
	c.Abort()
}
