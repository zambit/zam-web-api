package auth

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/ZamzamTech/wallet-api/db"
	"gitlab.com/ZamzamTech/wallet-api/models"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/base"
	"gitlab.com/ZamzamTech/wallet-api/services/sessions"
	"time"
	"gitlab.com/ZamzamTech/wallet-api/server/middlewares"
	"github.com/pkg/errors"
)

// SigninHandlerFactory returns handler which perform user authorization, requires session storage to store newly
// created session
func SigninHandlerFactory(
	d *db.Db,
	sessStorage sessions.IStorage,
	authExpiration time.Duration,
) base.HandlerFunc {
	wrongUserOrPasswordErr := base.NewErrorsView("wrong authorization data").AddField(
		"body", "phone", "either phone or password are invalid",
	)

	return func(c *gin.Context) (resp interface{}, code int, err error) {
		params := UserSigninRequest{}
		err = c.ShouldBindJSON(&params)
		if err != nil {
			return
		}

		// attempt to find user
		user, err := models.GetUserByPhoneAndStatus(d, params.Phone, models.UserStatusActive)
		if err != nil {
			if err == models.ErrUserNotFound {
				err = wrongUserOrPasswordErr
			}
			return
		}

		// compare hashed password with given
		passEqual, err := user.Password.Compare(params.Password)
		if err != nil {
			return
		}
		if !passEqual {
			err = wrongUserOrPasswordErr
			return
		}

		// create new session token
		token, err := sessStorage.New(map[string]interface{}{
			"id":    user.ID,
			"phone": string(user.Phone),
		}, authExpiration)
		if err != nil {
			// token not created, so whole handler failed
			return
		}

		// all is ok, auth has been passed
		resp = UserTokenResponse{Token: string(token)}
		return
	}
}

// SignoutHandlerFactory returns signout handler
func SignoutHandlerFactory(sessStorage sessions.IStorage, tokenName string) base.HandlerFunc {
	return func(c *gin.Context) (resp interface{}, code int, err error) {
		authToken, err := middlewares.GetAuthTokenFromContext(c, tokenName)
		if err != nil {
			return
		}

		err = sessStorage.Delete(sessions.Token(authToken))
		if err == sessions.ErrNotFound || err == sessions.ErrExpired {
			// shadow token miss to prevent token brute
			err = nil
		}
		return
	}
}

// RefreshTokenHandlerFactory returns handler which checks current token then refresh it
func RefreshTokenHandlerFactory(
	sessStorage sessions.IStorage,
	tokenName string,
	authExpiration time.Duration,
) base.HandlerFunc {

	return func(c *gin.Context) (resp interface{}, code int, err error) {
		authToken, err := middlewares.GetAuthTokenFromContext(c, tokenName)
		if err != nil {
			return
		}

		newToken, err := sessStorage.RefreshToken(sessions.Token(authToken), authExpiration)
		if err != nil {
			return
		}

		resp = UserTokenResponse{Token: string(newToken)}
		return
	}
}

// CheckHandlerFactory returns handler which returns user auth checking endpoint
func CheckHandlerFactory() base.HandlerFunc {
	return func(c *gin.Context) (resp interface{}, code int, err error) {
		userData := middlewares.GetUserDataFromContext(c)
		if userData == nil {
			err = errors.New("auth passed but no user data attached")
			return
		}
		resp = UserPhoneResponse{Phone: userData["phone"].(string)}
		return
	}
}