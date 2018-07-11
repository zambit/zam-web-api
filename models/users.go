package models

import (
	"gitlab.com/ZamzamTech/wallet-api/models/types"
	"time"
)

// User represents user model
type User struct {
	ID       int64
	Phone    types.Phone
	Password types.Password

	RegisteredAt time.Time

	ReferrerID    *int64
	ReferrerPhone string

	StatusID int64
	Status   UserStatusName
}

// NewUser creates new user from raw phone and password, also validates given fields
func NewUser(phone, password string, status UserStatusName, referrerPhone string) (user User, err error) {
	// create password
	encryptedPass, err := types.NewPass(password)
	if err != nil {
		return
	}

	// create phone
	parsedPhone, err := types.NewPhone(phone)
	if err != nil {
		return
	}

	user.Password = encryptedPass
	user.Phone = parsedPhone
	user.RegisteredAt = time.Now().UTC()
	user.Status = status
	user.ReferrerPhone = referrerPhone

	return
}

// UserStatusName represents UserStatuses table column type
type UserStatusName string

// Common user status names
const (
	UserStatusPending = UserStatusName("pending")
	UserStatusActive  = UserStatusName("active")
)

// UserStatus represents user status
type UserStatus struct {
	ID   int64
	Name UserStatusName
}
