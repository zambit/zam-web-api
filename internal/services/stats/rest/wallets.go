package rest

import (
	"encoding/json"
	"fmt"
	"git.zam.io/wallet-backend/common/pkg/merrors"
	"git.zam.io/wallet-backend/common/pkg/types"
	"git.zam.io/wallet-backend/common/pkg/types/decimal"
	"git.zam.io/wallet-backend/web-api/internal/services/stats"
	"git.zam.io/wallet-backend/web-api/pkg/server/handlers/base"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
)

// HTTPDoer exposes http client do method
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// UserWalletsGetter is IUserWalletsGetter which uses internal REST api of wallet-api service
type UserWalletsGetter struct {
	walletHost  string
	accessToken string
	client      HTTPDoer
}

// New UserWalletsGetter, also validates host and access token.
func New(host, accessToken string) (stats.IUserWalletsGetter, error) {
	_, err := url.Parse(host)
	if err != nil {
		return nil, errors.Wrap(err, "user wallets getter: wrong host parameter")
	}
	if accessToken == "" {
		return nil, errors.Wrap(err, "user wallets getter: access token is empty")
	}

	return &UserWalletsGetter{walletHost: host, accessToken: accessToken, client: &http.Client{}}, nil
}

type userStatResponse struct {
	Result bool `json:"result"`
	Data   struct {
		Count        int                      `json:"count"`
		TotalBalance map[string]*decimal.View `json:"total_balance"`
	} `json:"data"`
	Errors []base.ErrorView `json:"errors"`
}

const (
	userStatPath        = "/api/v1/internal/user_stat"
	defaultFiatCurrency = "usd"
)

// Get implements IUserWalletsGetter
func (g *UserWalletsGetter) Get(userPhone types.Phone, additionalFiatCurrency string) (
	stat stats.UserWalletsStats, err error,
) {
	if additionalFiatCurrency == "" {
		additionalFiatCurrency = defaultFiatCurrency
	}

	// do wallet-api stat request
	u, err := url.Parse(g.walletHost)
	if err != nil {
		err = errors.Wrap(err, "user wallets getter: wallet-api host is invalid")
		return
	}
	u.Path = userStatPath

	queryParams := make(url.Values)
	queryParams.Set("user_phone", string(userPhone))
	if additionalFiatCurrency != "" {
		queryParams.Set("convert", additionalFiatCurrency)
	}
	u.RawQuery = queryParams.Encode()

	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.accessToken))
	res, err := g.client.Do(req)
	if err != nil {
		err = errors.Wrap(err, "user wallets getter: wallet-api request failed")
		return
	}

	respBody := userStatResponse{}
	err = json.NewDecoder(res.Body).Decode(&respBody)
	if err != nil {
		err = errors.Wrap(err, "user wallets getter: wallet-api response decode failed")
		return
	}
	if !respBody.Result {
		for _, e := range respBody.Errors {
			err = merrors.Append(err, e)
		}
		err = errors.Wrap(err, "user wallets getter: rest api returns error")
		return
	}

	stat = stats.UserWalletsStats{
		Count:        respBody.Data.Count,
		TotalBalance: respBody.Data.TotalBalance,
	}
	return
}
