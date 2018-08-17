package stats

import (
	"git.zam.io/wallet-backend/common/pkg/types"
	"git.zam.io/wallet-backend/common/pkg/types/decimal"
	"github.com/pkg/errors"
)

// UserWalletsStats describes user wallets statistic
type UserWalletsStats struct {
	Count        int
	TotalBalance map[string]*decimal.View
}

// ErrInvalidFiatCurrency
var ErrInvalidFiatCurrency = errors.New("user wallet stats: invalid fiat currency")

// IUserWallets used to access user wallets statistic such as number of user wallets and total balances represented
// in BTC (default so far) and additional fiat currency
type IUserWalletsGetter interface {
	// Get stats. Fiat currency should be in short 3-letter form (example: USD, RUB, EUR etc), returns
	// ErrInvalidFiatCurrency in case of invalid value. Empty value forces default system fiat currency.
	Get(userPhone types.Phone, additionalFiatCurrency string) (UserWalletsStats, error)
}
