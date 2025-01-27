package recovery

import (
	"fmt"
	"git.zam.io/wallet-backend/common/pkg/merrors"
	"git.zam.io/wallet-backend/common/pkg/types"
	"git.zam.io/wallet-backend/web-api/db"
	models "git.zam.io/wallet-backend/web-api/internal/models/user"
	confflow "git.zam.io/wallet-backend/web-api/internal/server/handlers/flows/confirmation"
	"git.zam.io/wallet-backend/web-api/internal/services/isc"
	"git.zam.io/wallet-backend/web-api/internal/services/notifications"
	"git.zam.io/wallet-backend/web-api/pkg/server/handlers/base"
	"git.zam.io/wallet-backend/web-api/pkg/services/nosql"
	"github.com/pkg/errors"
	"time"
)

var (
	errFieldUserNotFound = base.NewFieldErr("body", "phone", "user not found")
)

const (
	verificationCodeKeyPattern = "user:%s:recovery:code"
	tokenKeyPattern            = "user:%s:recovery:token"
	notifSendTOKeyPattern      = "user:%s:recovery:notif_to"
)

func verificationCodeKey(user models.User) string {
	return fmt.Sprintf(verificationCodeKeyPattern, user.Phone)
}

func tokenKeyFunc(user models.User) string {
	return fmt.Sprintf(tokenKeyPattern, user.Phone)
}

func getUserState(_ db.ITx, storage nosql.IStorage, user models.User) (state confflow.State, err error) {
	codePresent, tokenPresent := false, false

	// check verification code first
	_, err = storage.Get(verificationCodeKey(user))
	if err != nil && err != nosql.ErrNoSuchKeyFound {
		return
	}
	codePresent = err != nosql.ErrNoSuchKeyFound
	err = nil

	// check token
	_, err = storage.Get(tokenKeyFunc(user))
	if err != nil && err != nosql.ErrNoSuchKeyFound {
		return
	}
	tokenPresent = err != nosql.ErrNoSuchKeyFound
	err = nil

	// estimate state
	switch {
	case codePresent && !tokenPresent:
		state = confflow.StatePending
	case !codePresent && tokenPresent:
		state = confflow.StateVerified
	case !codePresent && !tokenPresent:
		state = confflow.StateFinished
	case codePresent && tokenPresent:
		// inconsistent nosql storage state
		err = errors.New("inconsistent storage state: both recovery code and token present")
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
	notifSendTO time.Duration,
) base.HandlerFunc {
	resources := confflow.ExternalResources{
		Database:  d,
		Storage:   storage,
		Generator: generator,
	}
	return confflow.StartHandlerFactory(
		resources,
		func() interface{} {
			return &StartRequest{}
		},
		func(tx db.ITx, request interface{}) (user models.User, err error) {
			params := request.(*StartRequest)
			user, err = models.GetUserByPhoneAndStatus(tx, params.Phone, models.UserStatusActive, true)
			if err == models.ErrUserNotFound {
				err = errFieldUserNotFound
			}
			return
		},
		getUserState,
		func(tx db.ITx, storage nosql.IStorage, user models.User, newState confflow.State, params interface{}) (err error) {
			// do nothing, confirmation flow does all job for us
			return
		},
		func(resources confflow.ExternalResources, request interface{}, fErr error) (err error) {
			return postValidateFailedParams(d, fErr, request.(*StartRequest).Phone)
		},
		storageExpire,
		verificationCodeKeyPattern,
		func(user models.User, code string) error {
			return notifier.PasswordRecoveryVerificationRequested(fmt.Sprint(user.ID), string(user.Phone), code)
		},
		notifSendTO,
		notifSendTOKeyPattern,
		tokenKeyPattern,
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
			user, err = models.GetUserByPhoneAndStatus(tx, params.Phone, models.UserStatusActive, true)
			if err == models.ErrUserNotFound {
				err = errFieldUserNotFound
			}
			return
		},
		getUserState,
		func(tx db.ITx, storage nosql.IStorage, user models.User, newState confflow.State, params interface{}) (err error) {
			// do nothing, confirmation flow does all job for us
			return
		},
		func(resources confflow.ExternalResources, request interface{}, fErr error) (err error) {
			return postValidateFailedParams(d, fErr, request.(*VerifyRequest).Phone)
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
		tokenKeyPattern,
		storageExpire,
	)
}

// FinishHandlerFactory
func FinishHandlerFactory(d *db.Db, storage nosql.IStorage, notifier isc.IEventNotificator) base.HandlerFunc {
	resources := confflow.ExternalResources{
		Database: d,
		Storage:  storage,
	}

	return confflow.FinishHandlerFactory(
		resources,
		func() interface{} {
			return &FinishRequest{}
		},
		func(tx db.ITx, request interface{}) (user models.User, err error) {
			params := request.(*FinishRequest)
			user, err = models.GetUserByPhoneAndStatus(tx, params.Phone, models.UserStatusActive, true)
			if err == models.ErrUserNotFound {
				err = errFieldUserNotFound
			}
			return
		},
		getUserState,
		func(tx db.ITx, storage nosql.IStorage, user models.User, newState confflow.State, params interface{}) error {
			// parse pass
			password, err := types.NewPass(params.(*FinishRequest).Password)
			if err != nil {
				return err
			}

			// update user fields
			user.Password = password
			return models.UpdateUser(tx, user)
		},
		func(resources confflow.ExternalResources, request interface{}, fErr error) (err error) {
			return postValidateFailedParams(d, fErr, request.(*FinishRequest).Phone)
		},
		func(params interface{}) string {
			return params.(*FinishRequest).Token
		},
		func(tx db.ITx, user models.User) (resp interface{}, err error) {
			return
		},
		func(user models.User) error {
			return notifier.PasswordRecoveryCompleted(fmt.Sprint(user.ID), string(user.Phone))
		},
		tokenKeyPattern,
		"recovery_token",
	)
}

// utils
func postValidateFailedParams(d *db.Db, fErr error, phone string) (err error) {
	if !base.HaveFieldErr(fErr, "phone") && phone != "" {
		_, err = models.GetUserByPhone(d, phone)
		if err == models.ErrUserNotFound {
			fErr = merrors.Append(fErr, errFieldUserNotFound)
		}
	}
	if err != nil {
		return
	}
	return fErr
}
