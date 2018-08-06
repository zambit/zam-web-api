package signup

import (
	"fmt"
	"git.zam.io/wallet-backend/web-api/db"
	"git.zam.io/wallet-backend/web-api/internal/models"
	"git.zam.io/wallet-backend/web-api/internal/models/types"
	"git.zam.io/wallet-backend/web-api/pkg/server/handlers/base"
	confflow "git.zam.io/wallet-backend/web-api/internal/server/handlers/flows/confirmation"
	"git.zam.io/wallet-backend/web-api/pkg/services/nosql"
	"git.zam.io/wallet-backend/web-api/internal/services/notifications"
	"git.zam.io/wallet-backend/web-api/pkg/services/sessions"
	"time"
	"git.zam.io/wallet-backend/web-api/internal/services/isc"
	"git.zam.io/wallet-backend/common/pkg/merrors"
)

var (
	errFieldUserAlreadyExists = base.NewFieldErr("body", "phone", "user already exists")
	errFieldUserNotFound = base.NewFieldErr("body", "phone", "user not found")
	errFieldReferrerNotFound = base.NewFieldErr("body", "referrer_phone", "referrer not found")
)

const (
	verificationCodeKeyPattern = "user:%s:signup:code"
	signupTokenKeyPatten       = "user:%s:signup:token"
)

func getUserState(tx db.ITx, storage nosql.IStorage, user models.User) (state confflow.State, err error) {
	switch user.Status {
	case models.UserStatusCreated:
		state = confflow.StatePending
	case models.UserStatusPending:
		state = confflow.StatePending
	case models.UserStatusVerified:
		state = confflow.StateVerified
	case models.UserStatusActive:
		state = confflow.StateFinished
	default:
		err = fmt.Errorf("unexpected user status %s occured", user.Status)
	}
	return
}

// StartHandlerFactory
func StartHandlerFactory(
	d *db.Db,
	notifier isc.IEventNotificator,
	generator notifications.IGenerator,
	storage nosql.IStorage,
	storageExpire time.Duration,
) base.HandlerFunc {
	resources := confflow.ExternalResources{
		Database:    d,
		Storage:     storage,
		Generator:   generator,
	}
	return confflow.StartHandlerFactory(
		resources,
		func() interface{} {
			return &StartRequest{}
		},
		func(tx db.ITx, request interface{}) (user models.User, err error) {
			params := request.(*StartRequest)

			// fetch user by given phone
			user, err = models.GetUserByPhone(tx, params.Phone, true)
			if err != nil {
				// if no such phone registered we will create user with "crated" status
				if err == models.ErrUserNotFound {
					user, err = models.NewUser(params.Phone, "", models.UserStatusCreated, &params.ReferrerPhone)
					if err != nil {
						// seems that validator was failed, return internal error in such case
						return
					}

					// unique phone constraint will prevent concurrent creation (call will holds until first tx
					// will commit (in this case ErrUserAlreadyExists will be raised) or rollback changes
					user, err = models.CreateUser(tx, user)
					if err != nil {
						if err == models.ErrReferrerNotFound {
							err = errFieldReferrerNotFound
						}
						return
					}
				} else {
					return
				}
			}

			// not allowed in active state
			if user.Status == models.UserStatusActive {
				err = errFieldUserAlreadyExists
			}
			return
		},
		getUserState,
		func(tx db.ITx, storage nosql.IStorage, user models.User, newState confflow.State, params interface{}) (err error) {
			// update user status even if it remains unchanged
			// all returned errors, even logical, treated as internal
			_, err = models.UpdateUserStatus(tx, user, models.UserStatusPending)
			return
		},
		func(resources confflow.ExternalResources, request interface{}, fErr error) error {
			params := request.(*StartRequest)

			if !base.HaveFieldErr(fErr, "phone") {
				_, err := models.GetUserByPhone(d, params.Phone)
				if err == nil {
					fErr = merrors.Append(fErr, errFieldUserAlreadyExists)
				} else if err != models.ErrUserNotFound {
					return err
				}
			}

			if params.ReferrerPhone != "" && !base.HaveFieldErr(fErr, "referrer_phone") {
				_, err := models.GetUserByPhone(d, params.ReferrerPhone)
				if err == models.ErrUserNotFound {
					fErr = merrors.Append(fErr, errFieldReferrerNotFound)
				} else if err != nil {
					return err
				}
			}

			return fErr
		},
		storageExpire,
		verificationCodeKeyPattern,
		func(userID, userPhone, code string) error {
			return notifier.RegistrationVerificationRequested(userID, userPhone, code)
		},
		signupTokenKeyPatten,
	)
}

// VerifyHandlerFactory
func VerifyHandlerFactory(
	d *db.Db,
	generator notifications.IGenerator,
	storage nosql.IStorage,
	storageExpire time.Duration,
) base.HandlerFunc {
	resources := confflow.ExternalResources{
		Database:  d,
		Storage:   storage,
		Generator: generator,
	}

	return confflow.VerifyHandlerFactory(
		resources,
		func() interface{} {
			return &VerifyRequest{}
		},
		func(tx db.ITx, request interface{}) (user models.User, err error) {
			params := request.(*VerifyRequest)
			return models.GetUserByPhone(tx, params.Phone, true)
		},
		getUserState,
		func(tx db.ITx, storage nosql.IStorage, user models.User, newState confflow.State, params interface{}) (err error) {
			// update user status
			_, err = models.UpdateUserStatus(tx, user, models.UserStatusVerified)
			return
		},
		func(resources confflow.ExternalResources, request interface{}, fErr error) (err error) {
			params := request.(*VerifyRequest)

			if !base.HaveFieldErr(fErr, "phone") && params.Phone != "" {
				_, err = models.GetUserByPhone(d, params.Phone)
				if err == models.ErrUserNotFound {
					fErr = merrors.Append(fErr, errFieldUserNotFound)
					err = nil
				}
			}
			if err == nil {
				return
			}
			return fErr
		},
		func(request interface{}) string {
			return request.(*VerifyRequest).Code
		},
		func(token string) interface{} {
			return TokenView{
				Token: token,
			}
		},
		verificationCodeKeyPattern,
		signupTokenKeyPatten,
		storageExpire,
	)
}

// FinishHandlerFactory
func FinishHandlerFactory(
	d *db.Db,
	storage nosql.IStorage,
	notifier isc.IEventNotificator,
	sessStorage sessions.IStorage,
	authExpiration time.Duration,
) base.HandlerFunc {
	resources := confflow.ExternalResources{
		Database:    d,
		Storage:     storage,
	}

	return confflow.FinishHandlerFactory(
		resources,
		func() interface{} {
			return &FinishRequest{}
		},
		func(tx db.ITx, request interface{}) (user models.User, err error) {
			params := request.(*FinishRequest)
			return models.GetUserByPhone(tx, params.Phone, true)
		},
		getUserState,
		func(tx db.ITx, storage nosql.IStorage, user models.User, newState confflow.State, params interface{}) error {
			// update user fields
			// parse pass
			password, err := types.NewPass(params.(*FinishRequest).Password)
			if err != nil {
				return err
			}

			user.Password = password
			user.Status = models.UserStatusActive
			return models.UpdateUser(tx, user)
		},
		func(resources confflow.ExternalResources, request interface{}, fErr error) (err error) {
			params := request.(*FinishRequest)

			// check logical errors
			if !base.HaveFieldErr(fErr, "phone") && params.Phone != "" {
				_, err = models.GetUserByPhone(d, params.Phone)
				if err == models.ErrUserNotFound {
					fErr = merrors.Append(fErr, errFieldUserNotFound)
					err = nil
				}
			}
			if err == nil {
				return
			}
			return fErr
		},
		func(params interface{}) string {
			return params.(*FinishRequest).Token
		},
		func(tx db.ITx, user models.User) (resp interface{}, err error) {
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
		},
		func(userID string) error {
			return notifier.RegistrationCompleted(userID)
		},
		signupTokenKeyPatten,
		"signup_token",
	)
}
