package providers

import (
	"errors"
	"fmt"
	serverconf "git.zam.io/wallet-backend/web-api/config/server"
	"git.zam.io/wallet-backend/web-api/pkg/services/nosql"
	"git.zam.io/wallet-backend/web-api/pkg/services/sessions"
	"git.zam.io/wallet-backend/web-api/pkg/services/sessions/jwt"
	"git.zam.io/wallet-backend/web-api/pkg/services/sessions/mem"
	"time"
	"git.zam.io/wallet-backend/web-api/internal/services/notifications"
)

// SessionsStorage
func SessionsStorage(conf serverconf.Scheme, persistentStorage nosql.IStorage) (res sessions.IStorage, err error) {
	// catch jwt storage panics
	defer func() {
		r := recover()
		if r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				panic(r)
			}
		}
	}()

	switch conf.Auth.TokenStorage {
	case "mem", "":
		return mem.New(), nil
	case "jwt", "jwtpersistent":
		if conf.JWT == nil {
			return nil, errors.New("jwt like token storage required, but jwt configuration not provided")
		}
		res = jwt.New(conf.JWT.Method, []byte(conf.JWT.Secret), func() time.Time { return time.Now().UTC() })

		if conf.Auth.TokenStorage == "jwtpersistent" {
			res = jwt.WithStorage(
				res, persistentStorage, func(data map[string]interface{}, token string) string {
					return fmt.Sprintf("user:%v:sessions", data["phone"])
				},
			)
		}
		return
	default:
		return nil, fmt.Errorf("unsupported token storage type: %s", conf.Auth.TokenStorage)
	}
}

// Generator
func Generator(conf serverconf.Scheme) notifications.IGenerator {
	return notifications.NewWithCodeAlphabet(conf.Generator.CodeLen, conf.Generator.CodeAlphabet)
}