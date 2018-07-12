package cmd_test

import (
	"testing"

	"bytes"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gitlab.com/ZamzamTech/wallet-api/cmd/root"
	"gitlab.com/ZamzamTech/wallet-api/cmd/server"
	"gitlab.com/ZamzamTech/wallet-api/config"
	"time"
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Suite")
}

var exampleConfigStream = bytes.NewBuffer([]byte(
	`
env: example
server:
    host: "example.org"
    port: 1234

db:
    uri: postgresql://test:test@example.org/example
`,
))

var exampleArgs1 = []string{
	"-e", "sample",
	"-l", "sample.net",
	"-p", "4321",
	"--db.uri", "postgresql://sample:sample@sample.net/sample",
}

var exampleArgs2 = []string{
	"--env", "sample",
	"--server.host", "sample.net",
	"--server.port", "4321",
	"--db.uri", "postgresql://sample:sample@sample.net/sample",
}

var exampleEnc = map[string]string{
	"WA_ENV":         "poof",
	"WA_SERVER_HOST": "poof.org",
	"WA_SERVER_PORT": "5115",
	"WA_DB_URI":      "postgresql://poof:poof@poof.org/poof",
}

var _ = Describe("testing commands", func() {
	var v *viper.Viper
	var rootCmd cobra.Command
	var serverCmd cobra.Command
	BeforeEach(func() {
		v = viper.New()
		config.Init(v)
		rootCmd = root.Create(v, nil)
		serverCmd = server.Create(v, nil)
		rootCmd.AddCommand(&serverCmd)
	})

	Context("when using defaults (no config)", func() {
		It("should use defaulted db conn params", func() {
			conf := config.RootScheme{}

			err := v.Unmarshal(&conf)
			Expect(err).NotTo(HaveOccurred())

			Expect(conf.Env).To(Equal("test"))
			Expect(conf.DB.URI).To(Equal("postgresql://postgres:postgres@localhost:5432/postgres"))
			Expect(conf.Server.Host).To(Equal("localhost"))
			Expect(conf.Server.Port).To(Equal(9999))
			Expect(conf.Server.Auth.TokenName).To(Equal("Bearer"))
			Expect(conf.Server.Auth.TokenExpire).To(Equal(time.Hour*24))
		})
	})
	Context("when reading from config", func() {
		It("should read yaml like schema", func() {
			conf := config.RootScheme{}

			v.SetConfigType("yaml")
			err := v.ReadConfig(exampleConfigStream)
			Expect(err).NotTo(HaveOccurred())

			err = v.Unmarshal(&conf)
			Expect(err).NotTo(HaveOccurred())

			Expect(conf.Env).To(Equal("example"))
			Expect(conf.DB.URI).To(Equal("postgresql://test:test@example.org/example"))
			Expect(conf.Server.Host).To(Equal("example.org"))
			Expect(conf.Server.Port).To(Equal(1234))
		})
	})
	Context("when reading from args", func() {
		table.DescribeTable(
			"should bind passed args",
			func(args []string) {
				conf := config.RootScheme{}

				flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
				flags.AddFlagSet(serverCmd.Flags())
				flags.AddFlagSet(rootCmd.Flags())
				serverCmd.PersistentFlags()

				err := flags.Parse(args)
				Expect(err).NotTo(HaveOccurred())

				err = v.Unmarshal(&conf)
				Expect(err).NotTo(HaveOccurred())

				Expect(conf.Env).To(Equal("sample"))
				Expect(conf.DB.URI).To(Equal("postgresql://sample:sample@sample.net/sample"))
				Expect(conf.Server.Host).To(Equal("sample.net"))
				Expect(conf.Server.Port).To(Equal(4321))
			},
			table.Entry("use shortened args", exampleArgs1),
			table.Entry("use extended args", exampleArgs2),
		)
	})
})
