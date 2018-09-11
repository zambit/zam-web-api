package auth

import (
	"fmt"
	"git.zam.io/wallet-backend/web-api/db"
	models "git.zam.io/wallet-backend/web-api/internal/models/user"
	"git.zam.io/wallet-backend/web-api/internal/services/stats"
	"git.zam.io/wallet-backend/web-api/pkg/server/handlers/base"
	"git.zam.io/wallet-backend/web-api/pkg/server/middlewares"
	"git.zam.io/wallet-backend/web-api/pkg/services/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)

var errWrongUserOrPass = base.NewFieldErr("body", "phone", "either phone or password are invalid")

// SigninHandlerFactory returns handler which perform user authorization, requires session storage to store newly
// created session
func SigninHandlerFactory(
	d *db.Db,
	sessStorage sessions.IStorage,
	authExpiration time.Duration,
) base.HandlerFunc {
	return func(c *gin.Context) (resp interface{}, code int, err error) {
		params := UserSigninRequest{}
		err = base.ShouldBindJSON(c, &params)
		if err != nil {
			return
		}

		// attempt to find user
		user, err := models.GetUserByPhoneAndStatus(d, params.Phone, models.UserStatusActive)
		if err != nil {
			if err == models.ErrUserNotFound {
				err = errWrongUserOrPass
			}
			return
		}

		// compare hashed password with given
		passEqual, err := user.Password.Compare(params.Password)
		if err != nil {
			return
		}
		if !passEqual {
			err = errWrongUserOrPass
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
		phone, err := getUserPhone(c)
		if err != nil {
			return
		}
		resp = UserPhoneResponse{Phone: phone}
		return
	}
}

// StatFactory returns user statistic part of which is gathered from wallet api.
func StatFactory(d *db.Db, statsGetter stats.IUserWalletsGetter) base.HandlerFunc {
	return func(c *gin.Context) (resp interface{}, code int, err error) {
		// bind query params, ignore error
		params := UserMeRequest{}
		c.BindQuery(&params)

		phone, err := getUserPhone(c)
		if err != nil {
			return
		}

		//
		user, err := models.GetUserByPhone(d, phone)
		if err != nil {
			return
		}

		// query userStats
		userStats, err := statsGetter.Get(user.Phone, params.Convert)
		if err != nil {
			return
		}

		// prepare response
		resp = UserResponse{
			ID:           fmt.Sprint(user.ID),
			Phone:        phone,
			Status:       string(user.Status),
			RegisteredAt: user.RegisteredAt.Unix(),
			Wallets:      WalletsStatsView(userStats),
		}

		return
	}
}

// utils
func getUserPhone(c *gin.Context) (string, error) {
	userData := middlewares.GetUserDataFromContext(c)
	if userData == nil {
		return "", errors.New("auth passed but no user data attached")
	}
	return userData["phone"].(string), nil
}
