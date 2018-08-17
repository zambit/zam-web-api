package stats

import "github.com/ericlagergren/decimal"

//
type UserWalletsStats struct {
	Count        int
	TotalBalance map[string]*decimal.Big
}

// IUserStats
type IUserStats interface {
	Wallet
}
