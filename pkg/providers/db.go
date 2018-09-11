package providers

import (
	dbconf "git.zam.io/wallet-backend/web-api/config/db"
	"git.zam.io/wallet-backend/web-api/db"
)

// DB
func DB(conf dbconf.Scheme) (*db.Db, error) {
	return db.New(conf.URI)
}
