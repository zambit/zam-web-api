package fixtures

import (
	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"git.zam.io/wallet-backend/web-api/config"
	dbconfig "git.zam.io/wallet-backend/web-api/config/db"
	servconf "git.zam.io/wallet-backend/web-api/config/server"
	"go.uber.org/dig"
	"os"
	"time"
)

// Container global and used by test suites to fetch dependencies, available only inside test scope
var Container *dig.Container = nil

// beforeEachAssertErr shorthand for BeforeEach calling func and assert err not occurred
func beforeEachAssertErr(f func() error, timeout ...float64) bool {
	return ginkgo.BeforeEach(func() {
		if err := f(); err != nil {
			panic(err)
		}
	}, timeout...)
}

// afterEachAssertErr shorthand for AfterEach calling func and assert err not occurred
func afterEachAssertErr(f func() error, timeout ...float64) bool {
	return ginkgo.AfterEach(func() {
		if err := f(); err != nil {
			panic(err)
		}
	}, timeout...)
}

// BeforeEachCProvide allows to provide some dependency before each test
func BeforeEachCProvide(f interface{}, timeout ...float64) bool {
	return beforeEachAssertErr(func() error {
		return Container.Provide(f)
	}, timeout...)
}

// AfterEachCProvide allows to provide some dependency after each test
func AfterEachCProvide(f interface{}, timeout ...float64) bool {
	return afterEachAssertErr(func() error {
		return Container.Provide(f)
	}, timeout...)
}

// BeforeEachCInvoke is like ginkgo.BeforeEach and dig.Container.Invoke joined
func BeforeEachCInvoke(f interface{}, timeout ...float64) bool {
	return beforeEachAssertErr(func() error {
		return Container.Invoke(f)
	}, timeout...)
}

// BeforeEachCInvoke is like ginkgo.AfterEach and dig.Container.Invoke joined
func AfterEachCInvoke(f interface{}, timeout ...float64) bool {
	return afterEachAssertErr(func() error {
		return Container.Invoke(f)
	}, timeout...)
}

//
func ItD(text string, f interface{}) {
	ginkgo.It(text, func() {
		Expect(Container.Invoke(f)).NotTo(HaveOccurred())
	})
}

// init put basic dependencies into container
func Init() {
	// create before each test and remove container after test
	ginkgo.BeforeEach(func() {
		Container = dig.New()
	})
	ginkgo.AfterEach(func() {
		Container = nil
	})

	// testing configuration
	conf := config.RootScheme{
		DB: dbconfig.Scheme{
			URI: "postgresql://test:test@localhost/test?sslmode=disable",
		},
		Server: servconf.Scheme{
			Auth: servconf.AuthScheme{TokenName: "Test", TokenExpire: time.Second},
		},
	}

	// lookup for env variable
	if dbUri, ok := os.LookupEnv("WA_DB_URI"); ok {
		conf.DB.URI = dbUri
	}

	// provide config values
	BeforeEachCProvide(func() config.RootScheme {
		return conf
	})

	// provide db part
	BeforeEachCProvide(func() dbconfig.Scheme {
		return conf.DB
	})
}
