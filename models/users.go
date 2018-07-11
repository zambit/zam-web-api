package models

import (
	"gitlab.com/ZamzamTech/wallet-api/models/types"
	"time"
)

// User represents user model
type User struct {
	ID       int64
	Phone    string
	Password types.Password

	RegisteredAt time.Time

	ReferrerID    int64
	ReferrerPhone string

	StatusID int64
	Status   UserStatusName
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
