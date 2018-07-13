package migrations

import (
	"gitlab.com/ZamzamTech/wallet-api/fixtures"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	dbconfig "gitlab.com/ZamzamTech/wallet-api/config/db"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/golang-migrate/migrate/database/postgres"
)

func newMigrate(uri string) (*migrate.Migrate, error) {
	return migrate.New("file://db/migrations", uri)
}

func Init()  {
	var m *migrate.Migrate

	fixtures.BeforeEachCInvoke(func(conf dbconfig.Scheme) (err error) {
		// runs migrations in native way
		m, err = newMigrate(conf.URI)
		if err != nil {
			return
		}

		err = m.Up()
		// it noting changed drop db and start from scratch
		if err == migrate.ErrNoChange {
			err = m.Drop()
			if err != nil {
				return
			}

			err = m.Up()
		}
		return
	})

	ginkgo.AfterEach(func() {
		Expect(m).NotTo(BeNil())
		defer m.Close()
		if err := m.Down(); err != nil {
			if err == migrate.ErrNoChange {
				return
			}
			Expect(err).NotTo(HaveOccurred())
		}
	})
}
