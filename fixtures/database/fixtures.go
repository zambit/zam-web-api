package database

import (
	"gitlab.com/ZamzamTech/wallet-api/fixtures"
	"gitlab.com/ZamzamTech/wallet-api/db"
	dbconfig "gitlab.com/ZamzamTech/wallet-api/config/db"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
