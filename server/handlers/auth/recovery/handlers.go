package recovery

import (
	"fmt"
	"github.com/pkg/errors"
	"gitlab.com/ZamzamTech/wallet-api/db"
	"gitlab.com/ZamzamTech/wallet-api/models"
	"gitlab.com/ZamzamTech/wallet-api/models/types"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/base"
	confflow "gitlab.com/ZamzamTech/wallet-api/server/handlers/flows/confirmation"
	"gitlab.com/ZamzamTech/wallet-api/services/nosql"
	"gitlab.com/ZamzamTech/wallet-api/services/notifications"
	"time"
)

var (
	errFieldUserNotFound = base.FieldErrorDescr{
		Name:    "phone",
		Input:   "body",
		Message: "user not found",
	}
)

const (
	verificationCodeKeyPattern = "user:%s:recovery:code"
	tokenKeyPattern            = "user:%s:recovery:token"
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
	notifier notifications.ISender,
	generator notifications.IGenerator,
	storage nosql.IStorage,
	storageExpire time.Duration,
) base.HandlerFunc {
	resources := confflow.ExternalResources{
		Database:    d,
		Storage:     storage,
		Notificator: notifier,
		Generator:   generator,
	}
	return confflow.StartHandlerFactory(
		resources,
		func() interface{} {
			return &StartRequest{}
		},
		func(tx db.ITx, request interface{}) (user models.User, err error) {
			params := request.(*StartRequest)
			return models.GetUserByPhone(tx, params.Phone, true)
		},
		getUserState,
		func(tx db.ITx, storage nosql.IStorage, user models.User, newState confflow.State, params interface{}) (err error) {
			// do nothing, confirmation flow does all job for us
			return
		},
		func(resources confflow.ExternalResources, request interface{}, bErr base.FieldsErrorView) (err error) {
			return postValidateFailedParams(d, bErr, request.(*StartRequest).Phone)
		},
		storageExpire,
		verificationCodeKeyPattern,
		notifications.ActionPasswordRecoveryConfirmationRequested,
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
			return models.GetUserByPhone(tx, params.Phone, true)
		},
		getUserState,
		func(tx db.ITx, storage nosql.IStorage, user models.User, newState confflow.State, params interface{}) (err error) {
			// update user status
			_, err = models.UpdateUserStatus(tx, user, models.UserStatusVerified)
			return
		},
		func(resources confflow.ExternalResources, request interface{}, bErr base.FieldsErrorView) (err error) {
			return postValidateFailedParams(d, bErr, request.(*VerifyRequest).Phone)
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
func FinishHandlerFactory(d *db.Db, storage nosql.IStorage, notifier notifications.ISender) base.HandlerFunc {
	resources := confflow.ExternalResources{
		Database:    d,
		Storage:     storage,
		Notificator: notifier,
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
			// parse pass
			password, err := types.NewPass(params.(*FinishRequest).Password)
			if err != nil {
				return err
			}

			// update user fields
			user.Password = password
			return models.UpdateUser(tx, user)
		},
		func(resources confflow.ExternalResources, request interface{}, bErr base.FieldsErrorView) (err error) {
			return postValidateFailedParams(d, bErr, request.(*FinishRequest).Phone)
		},
		func(params interface{}) string {
			return params.(*FinishRequest).Token
		},
		func(tx db.ITx, user models.User) (resp interface{}, err error) {
			return
		},
		notifications.ActionPasswordRecoveryCompleted,
		tokenKeyPattern,
		"recovery_token",
	)
}

// utils
func postValidateFailedParams(d *db.Db, bErr base.FieldsErrorView, phone string) (err error) {
	// check logical errors
	skipUserCheck := false
	for _, f := range bErr.Fields {
		if f.Name == "phone" {
			skipUserCheck = true
		}
	}

	if !skipUserCheck {
		_, err = models.GetUserByPhone(d, phone)
		if err != nil {
			if err == models.ErrUserNotFound {
				err = nil
				bErr = bErr.AddFieldDescr(errFieldUserNotFound)
			}
		}
	}
	if err == nil {
		err = bErr
	}
	return
}
