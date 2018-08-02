package signup

import (
	"fmt"
	"git.zam.io/wallet-backend/web-api/db"
	"git.zam.io/wallet-backend/web-api/internal/models"
	"git.zam.io/wallet-backend/web-api/internal/server/handlers/base"
	"git.zam.io/wallet-backend/web-api/pkg/services/nosql"
	"git.zam.io/wallet-backend/web-api/internal/services/notifications"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"net/http"
	"time"
)

// sendUserConfirmationCode
func sendUserConfirmationCode(
	user models.User,
	generator notifications.IGenerator,
	notifier notifications.ISender,
	storage nosql.IStorage,
	storageExpire time.Duration,
) error {
	// issue new confirmation code
	code := generator.RandomCode()

	// save it in storage
	err := storage.SetWithExpire(confirmationCodeKey(user), code, storageExpire)
	if err != nil {
		return err
	}

	// send confirmation code
	return notifier.Send(
		notifications.ActionRegistrationConfirmationRequested,
		map[string]interface{}{
			"phone": string(user.Phone),
			"code":  code,
		},
		notifications.Confirmation,
	)
}

// confirmationCodeKey creates convient nosql storage key for confirmation code for specific user
func confirmationCodeKey(user models.User) string {
	return fmt.Sprintf("user_reg_conf_%s", user.Phone)
}

// signUpTokenKey creates convient nosql storage key for signup token for specific user
func signUpTokenKey(user models.User) string {
	return fmt.Sprintf("user_su_token_%s", user.Phone)
}

// ShouldBindJson
func ShouldBindJSON(c *gin.Context, to interface{}) (base.FieldsErrorView, error) {
	err := c.ShouldBindJSON(to)
	if err != nil {
		if vErr, ok := err.(validator.ValidationErrors); ok {
			return base.NewFieldsErrorsView(vErr), vErr
		}

		return base.FieldsErrorView{
			ErrorView: base.ErrorView{
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			},
		}, err
	}
	return base.FieldsErrorView{}, nil
}

// checkUserExistsAddFieldErr
func checkUserExistsAddFieldErr(
	tx db.ITx,
	fieldsErr base.FieldsErrorView,
	userPhone string,
	expectErr error,
	addErr base.FieldErrorDescr,
) (fErr base.FieldsErrorView, err error) {
	_, err = models.GetUserByPhone(tx, userPhone)
	if err == expectErr {
		err = nil
		fieldsErr.Fields = append(fieldsErr.Fields, addErr)
	}
	return fieldsErr, err
}
