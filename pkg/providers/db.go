package providers

import (
	"git.zam.io/wallet-backend/web-api/db"
	dbconf "git.zam.io/wallet-backend/web-api/config/db"
)

// DB
func DB(conf dbconf.Scheme) (*db.Db, error) {
	return db.New(conf.URI)
}
