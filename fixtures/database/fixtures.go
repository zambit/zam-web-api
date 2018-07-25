package database

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	dbconfig "git.zam.io/wallet-backend/web-api/config/db"
	"git.zam.io/wallet-backend/web-api/db"
	"git.zam.io/wallet-backend/web-api/fixtures"
)

// init provides database connection
func Init() {
	var d *db.Db

	fixtures.BeforeEachCProvide(func(conf dbconfig.Scheme) (database *db.Db, err error) {
		d, err = db.New(conf.URI)
		if err != nil {
			return
		}
		database = d
		return
	})

	ginkgo.AfterEach(func() {
		Expect(d).NotTo(BeNil())

		Expect(d.Close()).NotTo(HaveOccurred())
	})
}
