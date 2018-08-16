package auth

import (
	"git.zam.io/wallet-backend/common/pkg/types/decimal"
	"time"
)

// UserTokenResponse represents user sigin and signup responses
type UserTokenResponse struct {
	Token string `json:"token"`
}

// UserPhoneResponse represents user auth check response
type UserPhoneResponse struct {
	Phone string `json:"phone"`
}

// UserResponse
type UserResponse struct {
	ID           string    `json:"id"`
	Phone        string    `json:"phone"`
	Status       string    `json:"status"`
	RegisteredAt time.Time `json:"registered_at"`
	Wallets      struct {
		Count        int                      `json:"count"`
		TotalBalance map[string]*decimal.View `json:"total_balance"`
	} `json:"wallets"`
}
