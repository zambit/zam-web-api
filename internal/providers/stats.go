package providers

import (
	"git.zam.io/wallet-backend/web-api/config/isc"
	"git.zam.io/wallet-backend/web-api/internal/services/stats"
	"git.zam.io/wallet-backend/web-api/internal/services/stats/rest"
)

// UserWalletStatsGetter provides rest user stats getter
func UserWalletStatsGetter(conf isc.Scheme) (stats.IUserWalletsGetter, error) {
	return rest.New(conf.WalletApiDiscovery.Host, conf.WalletApiDiscovery.AccessToken)
}
