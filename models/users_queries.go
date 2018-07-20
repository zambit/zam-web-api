package models

import (
	"database/sql"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gitlab.com/ZamzamTech/wallet-api/db"
	"gitlab.com/ZamzamTech/wallet-api/models/types"
	"strconv"
)

var (
	// ErrInvalidUserID returned by query functions when user id is invalid
	ErrInvalidUserID = errors.New("not valid user identifier")

	// ErrInvalidUserStatus returned by query when user status is invalid
	ErrInvalidUserStatus = errors.New("invalid user status")

	// ErrUserNotFound returned when no user for given params
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists indicates that user with such phone already exists
	ErrUserAlreadyExists = errors.New("user already exists")

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
	user, err = doUserQuery(tx, `u.id = $1`, false, intID)
	return
}

// GetUserByPhone get user by raw phone, if forUpdate specified appropriate sql statement will be generated
func GetUserByPhone(tx db.ITx, phone string, forUpdate ...bool) (user User, err error) {
	phoneFormatted, err := types.NewPhone(phone)
	if err != nil {
		return
	}

	forUpdateCoerced := len(forUpdate) > 0 && forUpdate[0]
	user, err = doUserQuery(tx, `u.phone = $1`, forUpdateCoerced, phoneFormatted)
	return
}

// GetUserByPhone get user by raw phone and status name
func GetUserByPhoneAndStatus(tx db.ITx, phone string, status UserStatusName) (user User, err error) {
	phoneFormatted, err := types.NewPhone(phone)
	if err != nil {
		return
	}
	user, err = doUserQuery(tx, `u.phone = $1 AND us.name = $2`, false, phoneFormatted, status)
	return
}

// CreateUser create user with given status returns new user representation
func CreateUser(tx db.ITx, user User) (newUser User, err error) {
	newUser = user

	// Request referrer ID if it requested
	var refPhone types.Phone
	if user.ReferrerPhone != nil {
		refPhone, err = types.NewPhone(*user.ReferrerPhone)
		if err != nil {
			err = ErrReferrerNotFound
			return
		}
	}

	// ensure status related fields are populated
	user.SetStatus(user.Status)

	// since there is conditional statement which changes internals of sql query
	const (
		insertStart = `INSERT INTO users 
			(phone, password, created_at, registered_at, updated_at, referrer_id, status_id)`
		insertAppend = `RETURNING id`
		selectStatus = `(SELECT id FROM user_statuses WHERE name = $1)`
	)

	var query string
	var queryArgs []interface{}
	// looks so wild, perform so fast
	if user.ReferrerPhone != nil {
		query = insertStart + ` SELECT $2, $3, $4, $5, $6, u.id, ` + selectStatus + ` 
			FROM 
				users u
			INNER JOIN user_statuses us ON u.status_id = us.id
			WHERE 
				u.phone = $7 and 
				us.name = $8` + insertAppend
		queryArgs = []interface{}{
			user.Status,
			user.Phone,
			user.Password,
			user.CreatedAt,
			user.RegisteredAt,
			user.UpdatedAt,
			string(refPhone),
			UserStatusActive,
		}
	} else {
		query = insertStart + `VALUES ($2, $3, $4, $5, $6, $7, ` + selectStatus + `) ` + insertAppend
		queryArgs = []interface{}{
			user.Status, user.Phone, user.Password, user.CreatedAt, user.RegisteredAt, user.UpdatedAt, nil,
		}
	}

	err = tx.QueryRow(query, queryArgs...).Scan(&newUser.ID)
	if err != nil {
		// zero inserted rows means that no referrer with such phone exists
		if err == sql.ErrNoRows {
			err = ErrReferrerNotFound
			return
		} else if pgErr, ok := err.(pq.PGError); ok {
			if pgErr.Get('n') == "users_phone_idx" {
				err = ErrUserAlreadyExists
			}
		}
	}
	return
}

// UpdateUser updates whole user row
func UpdateUser(tx db.ITx, user User) (err error) {
	user.SetStatus(user.Status)
	statusID, err := getUserStatusID(tx, user.Status)
	if err != nil {
		err = ErrInvalidUserStatus
		return
	}

	res, err := tx.Exec(
		`UPDATE users SET
            phone = $1,
			password = $2,
            registered_at = $3,
			created_at = $4,
			updated_at = $5,
			status_id = $6
        WHERE id = $7`,
		user.Phone, user.Password, user.RegisteredAt, user.CreatedAt, user.UpdatedAt, statusID, user.ID,
	)
	if err != nil {
		return
	}
	rows, err := res.RowsAffected()
	switch {
	case err != nil:
		return
	case rows == 0:
		err = ErrUserNotFound
		return
	}
	return nil
}

// UpdateUserStatus changes user status, you must pass user with valid ID field
func UpdateUserStatus(tx db.ITx, user User, status UserStatusName) (newUser User, err error) {
	statusID, err := getUserStatusID(tx, status)
	if err != nil {
		err = ErrInvalidUserStatus
		return
	}

	// update status and related fields
	user.SetStatus(status)

	res, err := tx.Exec(
		`UPDATE users SET 
			status_id = $1,
			registered_at = $2,
            updated_at = $3
		WHERE id = $4`,
		statusID,
		user.RegisteredAt,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return
	}
	rows, err := res.RowsAffected()
	switch {
	case err != nil:
		return
	case rows == 0:
		err = ErrUserNotFound
		return
	}

	newUser = user
	newUser.StatusID = statusID
	return
}

func getUserStatusID(tx db.ITx, status UserStatusName) (id int64, err error) {
	// TODO may be locally cached
	res := tx.QueryRow(`SELECT id FROM user_statuses WHERE name = $1`, status)

	err = res.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrInvalidUserStatus
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

func doUserQuery(tx db.ITx, filter string, forUpdate bool, args ...interface{}) (user User, err error) {
	var forUpdatePart string
	if forUpdate {
		forUpdatePart = "FOR UPDATE OF u"
	}

	query := `SELECT
				u.id, u.phone, u.password, u.registered_at, 
		     	u.referrer_id, u.status_id, us.name, ru.phone 
         FROM users u 
		 LEFT JOIN users ru ON u.referrer_id = ru.id
		 INNER JOIN user_statuses us ON u.status_id = us.id
		 WHERE ` + filter + ` ` + forUpdatePart
	row := tx.QueryRow(query, args...)

	err = row.Scan(
		&user.ID,
		&user.Phone,
		&user.Password,
		&user.RegisteredAt,
		&user.ReferrerID,
		&user.StatusID,
		&user.Status,
		&user.ReferrerPhone,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrUserNotFound
		}
		return
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
