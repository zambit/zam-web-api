package migrations

import (
	dbconfig "git.zam.io/wallet-backend/web-api/config/db"
	"git.zam.io/wallet-backend/web-api/fixtures"
	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"path"
	"runtime"
)

func newMigrate(uri string) (*migrate.Migrate, error) {
	var migrationsDir string

	if envVal, ok := os.LookupEnv("WA_MIGRATIONS_DIR"); ok {
		migrationsDir = envVal
	} else {
		// determine migration location relative to this file
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			panic("No caller information")
		}
		projectRoot := path.Clean(path.Join(path.Dir(filename), "..", "..", ".."))
		migrationsDir = path.Clean(path.Join(projectRoot, "db/migrations"))
	}

	return migrate.New("file://"+migrationsDir, uri)
}

func Init() {
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
