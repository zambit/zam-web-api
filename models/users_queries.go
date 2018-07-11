package models

import (
	"database/sql"
	"github.com/pkg/errors"
	"gitlab.com/ZamzamTech/wallet-api/db"
	"gitlab.com/ZamzamTech/wallet-api/models/types"
	"strconv"
	"time"
)

var (
	// ErrInvalidUserID returned by query functions when user id is invalid
	ErrInvalidUserID = errors.New("not valid user identifier")

	// ErrInvalidUserStatus returned by query when user status is invalid
	ErrInvalidUserStatus = errors.New("invalid user status")

	// ErrUserNotFound returned when no user for given params
	ErrUserNotFound = errors.New("can't find user for given params")

	// ErrReferrerNotFound returned when user creation attempt failed because of wrong referrer phone
	ErrReferrerNotFound = errors.New("referrer not found")
)

// GetUserPhoneByID
func GetUserPhoneByID(tx db.ITx, id int64) (phone string, err error) {
	row := tx.QueryRow(`SELECT phone FROM users WHERE id = $1`, id)
	err = row.Scan(&phone)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrUserNotFound
		}
	}
	return
}

// GetUserByID performs user search by id query by given id using current transaction
func GetUserByID(tx db.ITx, id string) (user User, err error) {
	intID, err := parseUserID(id)
	if err != nil {
		return
	}

	// perform query
	// TODO queries must be prepared
	user, err = doUserQuery(tx, `SELECT * FROM users WHERE users.id = $1`, intID)
	return
}

// GetUserByPhone get user by raw phone
func GetUserByPhone(tx db.ITx, phone string) (user User, err error) {
	phoneFormatted, err := types.NewPhone(phone)
	if err != nil {
		return
	}

	user, err = doUserQuery(tx, `SELECT * FROM users WHERE users.phone = $1`, phoneFormatted)
	return
}

// CreateUser create user with given status returns new user representation
func CreateUser(tx db.ITx, user User) (newUser User, err error) {
	// TODO may be locally cached
	statusID, err := getUserStatusID(tx, user.Status)
	if err != nil {
		err = ErrInvalidUserID
		return
	}

	var referrerID *int64
	// Request referrer ID if it requested
	if user.ReferrerPhone != "" {
		referrer, err := GetUserByPhone(tx, user.ReferrerPhone)
		if err != nil {
			if err == ErrUserNotFound {
				err = ErrReferrerNotFound
			}
			return User{}, err
		}
		referrerID = &referrer.ID
	}

	// populate created at field
	if user.RegisteredAt.IsZero() {
		user.RegisteredAt = time.Now().UTC()
	}

	_, err = tx.Exec(
		`INSERT INTO users (phone, password, registered_at, referrer_id, status_id) 
        VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		user.Phone, user.Password, user.RegisteredAt, referrerID, statusID,
	)
	if err != nil {
		return
	}
	newUser = user
	// newUser.ID, err = res.LastInsertId()
	return
}

// ChangeUserStatus changes user status, you must pass user with valid ID field
func ChangeUserStatus(tx db.ITx, user User, status UserStatusName) (newUser User, err error) {
	statusID, err := getUserStatusID(tx, status)
	if err != nil {
		err = ErrInvalidUserStatus
		return
	}

	res, err := tx.Exec(`UPDATE users SET status_id = $1 WHERE id = $2`, statusID, user.ID)
	if err != nil {
		return
	}
	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		// returns actual error first
		if err != nil {
			return
		}
		if rows == 0 {
			err = ErrUserNotFound
		}

		return
	}

	newUser = user
	newUser.StatusID = statusID
	return
}

func getUserStatusID(tx db.ITx, status UserStatusName) (id int64, err error) {
	res := tx.QueryRow(`SELECT id FROM user_statuses WHERE name = $1`, status)

	err = res.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrInvalidUserID
		}
	}
	return
}

func getUserStatus(tx db.ITx, id int64) (status UserStatusName, err error) {
	res := tx.QueryRow(`SELECT name FROM user_statuses WHERE id = $1`, id)
	err = res.Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrInvalidUserStatus
		}
	}
	return
}

func doUserQuery(tx db.ITx, query string, args ...interface{}) (user User, err error) {
	row := tx.QueryRow(query, args...)

	err = row.Scan(&user.ID, &user.Phone, &user.Password, &user.RegisteredAt, &user.ReferrerID, &user.StatusID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrUserNotFound
		}
		return
	}

	// query referrer if it present
	if user.ReferrerID != 0 {
		referrerPhone, rErr := GetUserPhoneByID(tx, user.ReferrerID)
		if rErr != nil {
			if rErr != ErrUserNotFound {
				err = rErr
				return
			} else {
				// nothing to do here, db consistency failed!
			}
		}
		user.ReferrerPhone = referrerPhone
	}

	// load user status
	status, err := getUserStatus(tx, user.StatusID)
	if err != nil {
		if err == ErrInvalidUserStatus {
			// TODO remove it later
			status = UserStatusName("INVALID")
			err = nil
			return
		}
		user.Status = status
	}

	return
}

// utils
func parseUserID(id string) (int64, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, ErrInvalidUserID
	}
	return intID, nil
}
