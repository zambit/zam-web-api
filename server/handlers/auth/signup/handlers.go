package signup

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gitlab.com/ZamzamTech/wallet-api/db"
	"gitlab.com/ZamzamTech/wallet-api/models"
	"gitlab.com/ZamzamTech/wallet-api/models/types"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/base"
	"gitlab.com/ZamzamTech/wallet-api/services/nosql"
	"gitlab.com/ZamzamTech/wallet-api/services/notifications"
	"gitlab.com/ZamzamTech/wallet-api/services/sessions"
	"net/http"
	"time"
)

var (
	errNotAllowed = base.ErrorView{
		Code:    http.StatusBadRequest,
		Message: "such action not allowed",
	}
	errFieldUserAlreadyExists = base.FieldErrorDescr{
		Name:    "phone",
		Input:   "body",
		Message: "user already exists",
	}
	errFieldUserNotFound = base.FieldErrorDescr{
		Name:    "phone",
		Input:   "body",
		Message: "user not found",
	}
	errFieldReferrerNotFound = base.FieldErrorDescr{
		Name:    "referrer_phone",
		Input:   "body",
		Message: "referrer not found",
	}
	errFieldWrongCode = base.FieldErrorDescr{
		Name:    "verification_code",
		Input:   "body",
		Message: "code is wrong",
	}
	errFieldWrongToken = base.FieldErrorDescr{
		Name:    "signup_token",
		Input:   "body",
		Message: "signup_token is wrong",
	}
)

// StartHandlerFactory
func StartHandlerFactory(
	d *db.Db,
	notifier notifications.ISender,
	generator notifications.IGenerator,
	storage nosql.IStorage,
	storageExpire time.Duration,
	codeRetryDelay time.Duration,
) base.HandlerFunc {
	return func(c *gin.Context) (resp interface{}, httpCode int, err error) {
		// validate incoming params
		var params StartRequest
		bErr, err := ShouldBindJSON(c, &params)
		if err != nil {
			err = nil

			// perform additional validation so if there validation error, we still can return logical errors
			skipUserExistsErr, skipReferrerExistsErr := false, false
			for _, f := range bErr.Fields {
				switch f.Name {
				case "phone":
					skipUserExistsErr = true
				case "referrer_phone":
					skipReferrerExistsErr = true
				}
			}

			if !skipUserExistsErr {
				bErr, err = checkUserExistsAddFieldErr(d, bErr, params.Phone, nil, errFieldUserAlreadyExists)
				if err == models.ErrUserNotFound {
					err = nil
				}
			}
			if err != nil {
				return
			}
			if !skipReferrerExistsErr && params.ReferrerPhone != "" {
				bErr, err = checkUserExistsAddFieldErr(
					d, bErr, params.ReferrerPhone, models.ErrUserNotFound, errFieldReferrerNotFound,
				)
			}
			if err == nil {
				err = bErr
			}

			return
		}

		// do all queries in transaction to prevent concurrent user access/creation using SELECT FOR UPDATE
		err = d.Tx(func(tx db.ITx) (err error) {
			// fetch user by given phone
			user, err := models.GetUserByPhone(tx, params.Phone, true)
			if err != nil {
				// if no such phone registered we will create user with "crated" status
				if err == models.ErrUserNotFound {
					user, err = models.NewUser(params.Phone, "", models.UserStatusCreated, &params.ReferrerPhone)
					if err != nil {
						// seems that validator was failed, return internal error in such case
						return err
					}

					// unique phone constraint will prevent concurrent creation (call will holds until first tx
					// will commit (in this case ErrUserAlreadyExists will be raised) or rollback changes
					user, err = models.CreateUser(tx, user)
					if err != nil {
						if err == models.ErrReferrerNotFound {
							err = base.NewErrorsView("").AddFieldDescr(errFieldReferrerNotFound)
						}
						return err
					}
				} else {
					return err
				}
			}

			// not allowed in active state
			if user.Status == models.UserStatusActive {
				err = base.NewErrorsView("").AddFieldDescr(errFieldUserAlreadyExists)
				return
			}

			// prevent sms spam
			if !user.UpdatedAt.IsZero() {
				retryTimeDiff := time.Now().UTC().Sub(user.UpdatedAt.Add(codeRetryDelay))
				if retryTimeDiff > 0 {
					err = base.NewErrorsView(
						fmt.Sprintf("Not so fast! Next code dispatch will be awaliable in %v ...", retryTimeDiff),
					)
					return
				}
			}

			// issue new confirmation code
			code := generator.RandomCode()

			// save it in storage
			err = storage.SetWithExpire(confirmationCodeKey(user), code, storageExpire)
			if err != nil {
				return
			}
			// clear su token
			err = storage.Delete(signUpTokenKey(user))
			if err != nil {
				if err == nosql.ErrNoSuchKeyFound {
					err = nil
				}
			}

			// send confirmation code
			err = notifier.Send(
				notifications.ActionRegistrationConfirmationRequested,
				map[string]interface{}{
					"phone": string(user.Phone),
					"code":  code,
				},
				notifications.Confirmation,
			)
			// sadly, but whole transaction should be rollbacked if notification sent fails
			if err != nil {
				return
			}

			// update user status even if it remains unchanged
			// all returned errors, even logical, treated as internal
			_, err = models.UpdateUserStatus(tx, user, models.UserStatusPending)

			return
		})
		return
	}
}

// VerifyHandlerFactory
func VerifyHandlerFactory(
	d *db.Db,
	generator notifications.IGenerator,
	storage nosql.IStorage,
	storageExpire time.Duration,
) base.HandlerFunc {
	return func(c *gin.Context) (resp interface{}, code int, err error) {
		// validate incoming params
		var params VerifyRequest
		bErr, err := ShouldBindJSON(c, &params)
		if err != nil {
			err = nil

			// check logical errors
			skipUserCheck := false
			for _, f := range bErr.Fields {
				if f.Name == "phone" {
					skipUserCheck = true
				}
			}

			if !skipUserCheck {
				bErr, err = checkUserExistsAddFieldErr(
					d, bErr, params.Phone, models.ErrUserNotFound, errFieldUserNotFound,
				)
			}
			if err == nil {
				err = bErr
			}
			return
		}

		err = d.Tx(func(tx db.ITx) (err error) {
			// select user for update preventing concurrent modifications
			user, err := models.GetUserByPhone(tx, params.Phone, true)
			if err != nil {
				if err == models.ErrUserNotFound {
					err = base.NewErrorsView("").AddFieldDescr(errFieldUserNotFound)
				}
				return
			}

			// validate passed confirmation code
			codeKey := confirmationCodeKey(user)
			code, err := storage.Get(codeKey)
			if err == nosql.ErrNoSuchKeyFound || code != params.Code {
				err = base.NewErrorsView("").AddFieldDescr(errFieldWrongCode)
				return
			} else if err != nil {
				return
			}

			// check state after code confirmation to prevent phones leaks
			if user.Status != models.UserStatusPending {
				err = errNotAllowed
				return
			}

			// generate new signup token
			token := generator.RandomToken()
			tokenKey := signUpTokenKey(user)
			err = storage.SetWithExpire(tokenKey, token, storageExpire)
			if err != nil {
				return
			}

			// update user status
			_, err = models.UpdateUserStatus(tx, user, models.UserStatusVerified)

			// prepare response
			resp = TokenView{
				Token: token,
			}

			return
		})
		return
	}
}

// FinishHandlerFactory
func FinishHandlerFactory(
	d *db.Db,
	storage nosql.IStorage,
	notifier notifications.ISender,
	sessStorage sessions.IStorage,
	authExpiration time.Duration,
) base.HandlerFunc {
	return func(c *gin.Context) (resp interface{}, code int, err error) {
		// validate incoming params
		var params FinishRequest
		bErr, err := ShouldBindJSON(c, &params)
		if err != nil {
			err = nil

			// check logical errors
			skipUserCheck := false
			for _, f := range bErr.Fields {
				if f.Name == "phone" {
					skipUserCheck = true
				}
			}

			if !skipUserCheck {
				bErr, err = checkUserExistsAddFieldErr(
					d, bErr, params.Phone, models.ErrUserNotFound, errFieldUserNotFound,
				)
			}
			if err == nil {
				err = bErr
			}
			return
		}

		err = d.Tx(func(tx db.ITx) (err error) {
			// select user for update preventing concurrent modifications
			user, err := models.GetUserByPhone(tx, params.Phone, true)
			if err != nil {
				if err == models.ErrUserNotFound {
					err = base.NewErrorsView("").AddFieldDescr(errFieldUserNotFound)
				}
				return
			}

			// validate token
			tokenKey := signUpTokenKey(user)
			token, err := storage.Get(tokenKey)
			if err == nosql.ErrNoSuchKeyFound || token != params.Token {
				err = base.NewErrorsView("").AddFieldDescr(errFieldWrongToken)
				return
			}

			// finish allowed only on verified state
			if user.Status != models.UserStatusVerified {
				err = errNotAllowed
				return
			} else if err != nil {
				return
			}

			// update user fields
			// parse pass
			password, err := types.NewPass(params.Password)
			if err != nil {
				return
			}

			user.Password = password
			user.Status = models.UserStatusActive
			err = models.UpdateUser(tx, user)
			if err != nil {
				return err
			}

			// notify successful registration
			err = notifier.Send(
				notifications.ActionRegistrationCompleted,
				map[string]interface{}{
					"id":    user.ID,
					"phone": user.Phone,
				},
				notifications.Urgent,
			)
			if err != nil {
				return
			}

			// generate auth token
			authToken, err := sessStorage.New(map[string]interface{}{
				"id":    user.ID,
				"phone": user.Phone,
			}, authExpiration)
			if err != nil {
				return
			}

			// prepare answer
			resp = FinishResponse{
				Token: string(authToken),
			}
			return
		})
		return
	}
}
