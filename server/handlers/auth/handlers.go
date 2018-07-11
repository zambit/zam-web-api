package auth

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/ZamzamTech/wallet-api/db"
	"gitlab.com/ZamzamTech/wallet-api/models"
	"gitlab.com/ZamzamTech/wallet-api/models/types"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/base"
	"gitlab.com/ZamzamTech/wallet-api/services/notifications"
	"gitlab.com/ZamzamTech/wallet-api/services/sessions"
	"time"
)

// SignupHandlerFactory returns linked with given values /auth/signup handler
func SignupHandlerFactory(
	d *db.Db,
	sessStorage sessions.IStorage,
	notificator notifications.ISender,
	authExpiration time.Duration,
) base.HandlerFunc {
	return func(c *gin.Context) (resp interface{}, code int, err error) {
		var params UserSignupRequest
		err = c.ShouldBindJSON(&params)
		if err != nil {
			return
		}

		// create user model
		referrerPhone := ""
		if params.ReferrerPhone != nil {
			referrerPhone = *params.ReferrerPhone
		}
		user, err := models.NewUser(params.Phone, params.Password, models.UserStatusPending, referrerPhone)
		if err != nil {
			// TODO now it's impossible to ensure phone, password and referrer phone is invalid, this must be
			// implemented inside validator!
			if err == types.ErrInvalidPhoneNumber {
				err = base.NewErrorsView("").AddField("body", "phone", err.Error())
			}
			return
		}

		// check user already exists
		_, err = models.GetUserByPhone(d, params.Phone)
		switch {
		case err == nil:
			// nil means that user already exists
			err = base.NewErrorsView("").AddField("body", "phone", "user already exists")
			return
		case err != models.ErrUserNotFound:
			return
		}

		// do it in transaction
		err = d.Tx(func(tx db.ITx) error {
			_, err = models.CreateUser(tx, user)
			return err
		})
		if err != nil {
			return
		}

		// create user auth token
		token, err := sessStorage.New(map[string]interface{}{
			"id":    user.ID,
			"phone": user.Phone,
		}, authExpiration)
		if err != nil {
			c.Error(err)
			err = nil
		}

		// send notification
		// TODO signup verification request is required, nice place to do it here
		err = notificator.Send(
			notifications.ActionRegistrationCompleted,
			map[string]interface{}{
				"phone": user.Phone,
			},
			notifications.Urgent,
		)
		if err != nil {
			c.Error(err)
			err = nil
		}

		resp, code = UserTokenResponse{
			Token: string(token),
		}, 201
		return
	}
}
