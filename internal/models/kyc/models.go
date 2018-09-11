package kyc

import (
	"git.zam.io/wallet-backend/common/pkg/types/postgres"
	"time"
)

// StatusType
type StatusType string

const (
	StatusPending  StatusType = "pending"
	StatusVerified            = "verified"
	StatusDeclined            = "declined"
)

// PersonalData holds user personal information
type Data struct {
	ID     int64
	UserID int64

	Status   StatusType
	StatusID int64

	Email     string
	FirstName string
	LastName  string
	BirthDate time.Time
	Sex       string
	Country   string
	Address   postgres.JSONb
}
