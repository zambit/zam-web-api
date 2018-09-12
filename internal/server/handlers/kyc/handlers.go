package kyc

import (
	"git.zam.io/wallet-backend/web-api/db"
	"git.zam.io/wallet-backend/web-api/internal/models/kyc"
	"git.zam.io/wallet-backend/web-api/pkg/server/handlers/base"
	"git.zam.io/wallet-backend/web-api/pkg/server/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

var (
	errDataAlreadySent = base.ErrorView{Code: http.StatusBadRequest, Message: "personal data already sent"}
	errMinorPerson     = base.NewFieldErr("body", "birth_date", "minor age")
	errInvalidGender   = base.NewFieldErr(
		"body",
		"sex",
		`invalid gender, must be on of "male", "female" or "undefined"`,
	)
)

// CreateFactory
func CreateFactory(d *db.Db) base.HandlerFunc {
	return func(c *gin.Context) (resp interface{}, code int, err error) {
		params := CreateRequest{}
		err = base.ShouldBindJSON(c, &params)
		if err != nil {
			return
		}

		// get user id from context
		// TODO need to improve user data info attached to the user session
		userID, err := getUserIDFromContext(c)
		if err != nil {
			return
		}

		// check user age
		diff := yearsDiff(time.Time(*params.BirthDate), time.Now())
		if diff < 18 {
			err = errMinorPerson
			return
		}

		err = d.Tx(func(tx db.ITx) error {
			_, err := kyc.Create(tx, &kyc.Data{
				UserID:    userID,
				Status:    kyc.StatusPending,
				Email:     params.Email,
				FirstName: params.FirstName,
				LastName:  params.LastName,
				BirthDate: time.Time(*params.BirthDate),
				Sex:       params.Sex,
				Country:   params.Country,
				Address: map[string]interface{}{
					"city":        params.City,
					"region":      params.Region,
					"street":      params.Street,
					"house":       params.House,
					"postal_code": params.PostalCode,
				},
			})
			return err
		})
		switch err {
		case kyc.ErrInvalidGender:
			err = errInvalidGender
		case kyc.ErrAlreadyExists:
			err = errDataAlreadySent
		}
		return
	}
}

// GetFactory
func GetFactory(d *db.Db) base.HandlerFunc {
	return func(c *gin.Context) (resp interface{}, code int, err error) {
		// get user id from context
		// TODO need to improve user data info attached to the user session
		userID, err := getUserIDFromContext(c)
		if err != nil {
			return
		}

		var data *kyc.Data
		err = d.Tx(func(tx db.ITx) error {
			var err error
			data, err = kyc.Get(tx, userID)
			// this error means that KYC hasn't been started, so show nothing in such case
			if err == kyc.ErrNoSuchUser {
				err = nil
			}
			return err
		})
		if err != nil {
			return
		}
		resp = CreateGetResponse(data)
		return
	}
}

func getUserIDFromContext(c *gin.Context) (id int64, err error) {
	var userID struct {
		ID int64
	}

	data := middlewares.GetUserDataFromContext(c)
	if data == nil {
		err = errors.New("kyc: user auth middleware is missing")
		return
	}
	err = mapstructure.Decode(data, &userID)
	if err != nil {
		err = errors.Wrap(err, "kyc")
	}
	id = userID.ID
	return
}

func yearsDiff(a, b time.Time) (year int) {
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}
	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	year = int(y2 - y1)
	month := int(M2 - M1)
	day := int(d2 - d1)

	// Normalize negative values
	if day < 0 {
		// days in month:
		t := time.Date(y1, M1, 32, 0, 0, 0, 0, time.UTC)
		day += 32 - t.Day()
		month--
	}
	if month < 0 {
		month += 12
		year--
	}

	return
}
