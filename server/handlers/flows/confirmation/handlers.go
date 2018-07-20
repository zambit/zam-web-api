package confirmation

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gitlab.com/ZamzamTech/wallet-api/db"
	"gitlab.com/ZamzamTech/wallet-api/models"
	"gitlab.com/ZamzamTech/wallet-api/server/handlers/base"
	"gitlab.com/ZamzamTech/wallet-api/services/nosql"
	"gitlab.com/ZamzamTech/wallet-api/services/notifications"
	"net/http"
	"time"
)

type ExternalResources struct {
	Database    *db.Db
	Storage     nosql.IStorage
	Notificator notifications.ISender
	Generator   notifications.IGenerator
}

type State string

const (
	StatePending  State = "state_pending"
	StateVerified       = "state_verified"
	StateFinished       = "state_finished"
)

var (
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
	errNotAllowed = base.ErrorView{
		Code:    http.StatusBadRequest,
		Message: "such action not allowed",
	}
)

type ParamsFactory func() interface{}

type PostValidateFieldsFunc func(resources ExternalResources, request interface{}, err base.FieldsErrorView) error
type GetUserFunc func(tx db.ITx, request interface{}) (user models.User, err error)
type GetUserStateFunc func(tx db.ITx, storage nosql.IStorage, user models.User) (state State, err error)
type SetUserStateFunc func(tx db.ITx, storage nosql.IStorage, user models.User, newState State, params interface{}) error

// StartHandlerFactory creates
func StartHandlerFactory(
	resources ExternalResources,
	factory ParamsFactory,
	userFunc GetUserFunc,
	userStateFunc GetUserStateFunc,
	setUserStateFunc SetUserStateFunc,
	postValidateFunc PostValidateFieldsFunc,
	verifCodeExpire time.Duration,
	verifCodeKeyPattern string,
	verifCodeNotifAction string,
	finishTokenKeyPattern string,
) base.HandlerFunc {
	return func(c *gin.Context) (resp interface{}, code int, err error) {
		params, err := paramsOrErr(c, resources, factory, postValidateFunc)
		if err != nil {
			return
		}

		err = resources.Database.Tx(func(tx db.ITx) error {
			user, err := userFunc(tx, params)
			if err != nil {
				return err
			}

			_, err = userStateFunc(tx, resources.Storage, user)
			if err != nil {
				return err
			}

			// issue new confirmation code
			code := resources.Generator.RandomCode()

			// save it in storage
			err = resources.Storage.SetWithExpire(
				verificationCodeKey(verifCodeKeyPattern, user), code, verifCodeExpire,
			)
			if err != nil {
				return err
			}

			// clear finish token
			err = resources.Storage.Delete(finishTokenKey(finishTokenKeyPattern, user))
			if err != nil {
				if err == nosql.ErrNoSuchKeyFound {
					err = nil
				} else {
					return err
				}
			}

			// send confirmation code
			err = resources.Notificator.Send(
				verifCodeNotifAction,
				map[string]interface{}{
					"phone": string(user.Phone),
					"code":  code,
				},
				notifications.Confirmation,
			)

			// sadly, but whole transaction should be rollbacked if notification sent fails
			if err != nil {
				return err
			}

			// update state
			return setUserStateFunc(tx, resources.Storage, user, StatePending, params)
		})
		if err != nil {
			err = coerceFieldErr(err)
			return
		}
		return
	}
}

// VerifyHandlerFactory
func VerifyHandlerFactory(
	resources ExternalResources,
	factory ParamsFactory,
	userFunc GetUserFunc,
	userStateFunc GetUserStateFunc,
	setUserStateFunc SetUserStateFunc,
	postValidateFunc PostValidateFieldsFunc,
	getCodeFromParams func(interface{}) string,
	tokenRespFactory func(token string) interface{},
	verifCodeKeyPattern,
	finishTokenKeyPattern string,
	finishTokenExpire time.Duration,
) base.HandlerFunc {
	return func(c *gin.Context) (resp interface{}, code int, err error) {
		params, err := paramsOrErr(c, resources, factory, postValidateFunc)
		if err != nil {
			return
		}

		err = resources.Database.Tx(func(tx db.ITx) (err error) {
			user, err := userFunc(tx, params)
			if err != nil {
				return err
			}

			state, err := userStateFunc(tx, resources.Storage, user)
			if err != nil {
				return err
			}

			// validate passed verification code
			codeKey := verificationCodeKey(verifCodeKeyPattern, user)
			code, err := resources.Storage.Get(codeKey)
			if err == nosql.ErrNoSuchKeyFound || getCodeFromParams(params) != code {
				err = errFieldWrongCode
				return
			} else if err != nil {
				return
			}

			// remove verification code
			err = resources.Storage.Delete(codeKey)
			// skip no suck key found error
			if err != nil && err != nosql.ErrNoSuchKeyFound {
				return
			}

			// check state after code confirmation to prevent leaks
			if state != StatePending {
				err = errNotAllowed
				return
			}

			// generate new finish token
			token := resources.Generator.RandomToken()
			tokenKey := finishTokenKey(finishTokenKeyPattern, user)
			err = resources.Storage.SetWithExpire(tokenKey, token, finishTokenExpire)
			if err != nil {
				return
			}

			// update user status
			err = setUserStateFunc(tx, resources.Storage, user, StateVerified, params)
			if err != nil {
				return
			}

			// prepare response
			resp = tokenRespFactory(token)

			return
		})
		if err != nil {
			err = coerceFieldErr(err)
		}
		return
	}
}

// FinishHandlerFactory
func FinishHandlerFactory(
	resources ExternalResources,
	factory ParamsFactory,
	userFunc GetUserFunc,
	userStateFunc GetUserStateFunc,
	setUserStateFunc SetUserStateFunc,
	postValidateFunc PostValidateFieldsFunc,
	getTokenFromParams func(interface{}) string,
	respFactory func(tx db.ITx, user models.User) (interface{}, error),
	finishedAction string,
	finishTokenKeyPattern string,
	tokenFieldName string,
) base.HandlerFunc {
	return func(c *gin.Context) (resp interface{}, code int, err error) {
		params, err := paramsOrErr(c, resources, factory, postValidateFunc)
		if err != nil {
			return
		}

		err = resources.Database.Tx(func(tx db.ITx) (err error) {
			user, err := userFunc(tx, params)
			if err != nil {
				return err
			}

			state, err := userStateFunc(tx, resources.Storage, user)
			if err != nil {
				return err
			}

			// validate token
			tokenKey := finishTokenKey(finishTokenKeyPattern, user)
			token, err := resources.Storage.Get(tokenKey)
			if err == nosql.ErrNoSuchKeyFound || getTokenFromParams(params) != token {
				err = base.FieldErrorDescr{
					Name:    tokenFieldName,
					Input:   "body",
					Message: fmt.Sprintf("%s is wrong", tokenFieldName),
				}
				return
			}
			// delete finish token
			err = resources.Storage.Delete(tokenKey)
			if err != nil {
				return
			}

			// finish allowed only on verified state
			if state != StateVerified {
				err = errNotAllowed
				return
			} else if err != nil {
				return
			}

			// update user fields
			// parse pass
			err = setUserStateFunc(tx, resources.Storage, user, StateFinished, params)
			if err != nil {
				return
			}

			if finishedAction != "" {
				// notify about finish
				err = resources.Notificator.Send(
					finishedAction,
					map[string]interface{}{
						"id":    user.ID,
						"phone": user.Phone,
					},
					notifications.Urgent,
				)
				if err != nil {
					return
				}
			}

			// prepare answer
			resp, err = respFactory(tx, user)
			return
		})
		if err != nil {
			err = coerceFieldErr(err)
		}
		return
	}
}

// coerceErr
func coerceFieldErr(err error) error {
	switch e := err.(type) {
	case base.ErrorView:
		return err
	case base.FieldErrorDescr:
		return base.NewErrorsView("").AddFieldDescr(e)
	default:
		return e
	}
}

//
func paramsOrErr(
	c *gin.Context,
	resources ExternalResources,
	factory ParamsFactory,
	postValidateFunc PostValidateFieldsFunc,
) (params interface{}, err error) {
	params = factory()
	bErr, err := base.ShouldBindJSON(c, params)
	if err != nil {
		if len(bErr.Fields) == 0 || postValidateFunc == nil {
			return
		}
		err = postValidateFunc(resources, params, bErr)
		return
	}
	return
}

func verificationCodeKey(pattern string, user models.User) string {
	return fmt.Sprintf(pattern, user.Phone)
}

func finishTokenKey(pattern string, user models.User) string {
	return fmt.Sprintf(pattern, user.Phone)
}
