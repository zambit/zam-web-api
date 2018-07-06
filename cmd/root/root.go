package root

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Create and initialize root command for given viper instance
func Create(v *viper.Viper) cobra.Command {
	var config string

	command := cobra.Command{
		Use: "wallet-api",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if err = cmd.ParseFlags(args); err != nil {
				return
			}

			if config != "" {
				// Trying to open config
				v.SetConfigFile(config)

				// Attempts to load configuration
				err = v.ReadInConfig()
				if err != nil {
					return
				}
			}

			return nil
		},
	}

	command.Flags().StringVarP(
		&config, "config", "c", "", "specifies configuration file to load from",
	)
	command.Flags().StringP(
		"env", "e", "test", "specifies current environment (prod/dev/test)",
	)
	v.BindPFlags(command.Flags())

	return command
}