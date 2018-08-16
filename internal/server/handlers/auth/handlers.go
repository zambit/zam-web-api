package auth

import (
	"encoding/json"
	"fmt"
	"git.zam.io/wallet-backend/common/pkg/merrors"
	"git.zam.io/wallet-backend/common/pkg/types/decimal"
	"git.zam.io/wallet-backend/web-api/config/server"
	"git.zam.io/wallet-backend/web-api/db"
	"git.zam.io/wallet-backend/web-api/internal/models"
	"git.zam.io/wallet-backend/web-api/pkg/server/handlers/base"
	"git.zam.io/wallet-backend/web-api/pkg/server/middlewares"
	"git.zam.io/wallet-backend/web-api/pkg/services/sessions"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
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
		err = c.ShouldBindJSON(&params)
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
func StatFactory(d *db.Db, config server.Scheme) base.HandlerFunc {
	type userStatResponse struct {
		Result bool `json:"result"`
		Data   struct {
			Count        int                      `json:"count"`
			TotalBalance map[string]*decimal.View `json:"total_balance"`
		} `json:"data"`
		Errors []base.ErrorView `json:"errors"`
	}
	const userStatPath = "/api/v1/internal/user_stat"

	return func(c *gin.Context) (resp interface{}, code int, err error) {
		// bind query params, ignore error
		params := UserMeRequest{}
		c.BindQuery(&params)
		if params.Convert == "" {
			params.Convert = "usd"
		}

		phone, err := getUserPhone(c)
		if err != nil {
			return
		}

		//
		user, err := models.GetUserByPhone(d, phone)
		if err != nil {
			return
		}

		// do wallet-api stat request
		u, err := url.Parse(config.WalletApiDiscovery.Host)
		if err != nil {
			err = errors.Wrap(err, "handlers: user: wallet-api host is invalid")
			return
		}
		u.Path = userStatPath
		u.RawQuery = fmt.Sprintf("user_phone=%s&convert=%s", url.QueryEscape(phone), params.Convert)

		req, _ := http.NewRequest("GET", u.String(), nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.WalletApiDiscovery.AccessToken))
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			err = errors.Wrap(err, "handlers: user: wallet-api request failed")
			return
		}

		respBody := userStatResponse{}
		err = json.NewDecoder(res.Body).Decode(&respBody)
		if err != nil {
			err = errors.Wrap(err, "handlers: user: wallet-api response decode failed")
			return
		}
		if !respBody.Result {
			for _, e := range respBody.Errors {
				err = merrors.Append(err, e)
			}
			return
		}

		// prepare response
		resp = UserResponse{
			ID:           fmt.Sprint(user.ID),
			Phone:        phone,
			Status:       string(user.Status),
			RegisteredAt: *user.RegisteredAt,
			Wallets:      respBody.Data,
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
