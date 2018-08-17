package auth

import (
	"git.zam.io/wallet-backend/common/pkg/types/decimal"
)

// UserTokenResponse represents user sigin and signup responses
type UserTokenResponse struct {
	Token string `json:"token"`
}

// UserPhoneResponse represents user auth check response
type UserPhoneResponse struct {
	Phone string `json:"phone"`
}

// WalletsStatsView
type WalletsStatsView struct {
	Count        int                      `json:"count"`
	TotalBalance map[string]*decimal.View `json:"total_balance"`
}

// UserResponse
type UserResponse struct {
	ID           string           `json:"id"`
	Phone        string           `json:"phone"`
	Status       string           `json:"status"`
	RegisteredAt int64            `json:"registered_at"`
	Wallets      WalletsStatsView `json:"wallets"`
}
