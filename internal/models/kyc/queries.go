package kyc

import (
	"database/sql"
	"git.zam.io/wallet-backend/web-api/db"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

var (
	// ErrAlreadyExists
	ErrAlreadyExists = errors.New("kyc: data already created for this user")

	// ErrInvalidGender
	ErrInvalidGender = errors.New("kyc: invalid gender")

	// ErrInvalidStatus
	ErrInvalidStatus = errors.New("kyc: invalid status")

	// ErrNoSuchUser
	ErrNoSuchUser = errors.New("kyc: no such user")
)

// Create kyc data record
//
// Method doesn't require StatusID field to be filled and it uses Status name to lookup appropriate status id
func Create(tx db.ITx, data *Data) (id int64, err error) {
	err = tx.QueryRow(
		`insert into personal_data 
			(user_id, status_id, email, first_name, last_name, birth_date, sex, country, address)
		 values ($1, (select id from personal_data_statuses where name = $2), $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id`,
		data.UserID, data.Status, data.Email, data.FirstName, data.LastName,
		data.BirthDate, data.Sex, data.Country, data.Address,
	).Scan(&id)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch {
			case pqErr.Constraint == "personal_data_user_id_key":
				err = ErrAlreadyExists
			case pqErr.Code == "22P02":
				err = ErrInvalidGender
			case pqErr.Column == "user_id":
				err = ErrNoSuchUser
			case pqErr.Column == "status_id":
				err = ErrInvalidStatus
			}
		}
	}
	return
}

// Get get user kyc data by user id
func Get(tx db.ITx, userID int64) (data *Data, err error) {
	var d Data
	err = tx.QueryRow(
		`select
			id, user_id, 
			(select name from personal_data_statuses where id = status_id), status_id,
			email, first_name, last_name, birth_date, sex, country, address
         from personal_data where user_id = $1`,
		userID,
	).Scan(
		&d.ID,
		&d.UserID,
		&d.Status,
		&d.StatusID,
		&d.Email,
		&d.FirstName,
		&d.LastName,
		&d.BirthDate,
		&d.Sex,
		&d.Country,
		&d.Address,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrNoSuchUser
		}
		return
	}
	return &d, nil
}

// GetStatus get only status
func GetStatus(tx db.ITx, userID int64) (status StatusType, err error) {
	err = tx.QueryRow(
		`select 
			personal_data_statuses.name 
		from personal_data
		inner join personal_data_statuses on personal_data_statuses.id = personal_data.status_id
		where user_id = $1`,
		userID,
	).Scan(&status)
	return
}