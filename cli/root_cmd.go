package cli

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

const (
	PersistentFlagConfig = "config"
)

var (
	requiredFlags = []string{
		PersistentFlagConfig,
	}

	config = Config{}

	RootCmd = &cobra.Command{
		Use:           "nuki-logger",
		Short:         "Query Nuki api for most recent logs",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var requiredFlagsMissing []string
			for _, v := range requiredFlags {
				if viper.GetString(v) == "" {
					requiredFlagsMissing = append(requiredFlagsMissing, v)
				}
			}

			if len(requiredFlagsMissing) > 0 {
				return fmt.Errorf("the following flag(s) are required: %s", strings.Join(requiredFlagsMissing, ", "))
			}

			if viper.GetString(PersistentFlagConfig) != "" {
				viper.SetConfigName(viper.GetString(PersistentFlagConfig))
			}

			return config.LoadConfig(viper.GetViper())
		},
		RunE: run,
	}
)

func init() {
	RootCmd.PersistentFlags().StringP(PersistentFlagConfig, "c", "", "Configuration file")

	RootCmd.AddCommand(ServerCmd)
	RootCmd.AddCommand(QueryCmd)

	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.BindPFlags(RootCmd.PersistentFlags())
}

func run(_ *cobra.Command, _ []string) error {
	log.Debug().
		Msg(fmt.Sprintf("%+v", config))
	log.Error().
		Err(fmt.Errorf("please use one of the supported sub-command")).
		Send()
	return nil
}
