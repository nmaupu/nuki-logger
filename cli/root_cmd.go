package cli

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

const (
	EnvVarsPrefix             = "NUKI_LOGGER"
	PersistentFlagToken       = "token"
	PersistentFlagSmartlockID = "smartlockid"
)

var (
	requiredFlags = []string{
		PersistentFlagToken,
		PersistentFlagSmartlockID,
	}
	RootCmd = &cobra.Command{
		Use:           "nuki-logger",
		Short:         "Query Nuki api for most recent logs",
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var requiredFlagsMissing []string
			for _, v := range requiredFlags {
				if viper.GetString(v) == "" {
					requiredFlagsMissing = append(requiredFlagsMissing, v)
				}
			}

			if len(requiredFlagsMissing) > 0 {
				return fmt.Errorf("the following flag(s) are required: %s", strings.Join(requiredFlagsMissing, ", "))
			}

			return nil
		},
		RunE: run,
	}
)

func init() {
	RootCmd.PersistentFlags().StringP(PersistentFlagSmartlockID, "s", "", "Smartlock ID to get the logs from")
	RootCmd.PersistentFlags().StringP(PersistentFlagToken, "t", "", "Token to access the Nuki API")

	RootCmd.AddCommand(ServerCmd)
	RootCmd.AddCommand(QueryCmd)

	viper.AutomaticEnv()
	viper.SetEnvPrefix(EnvVarsPrefix)
	_ = viper.BindPFlags(RootCmd.PersistentFlags())
}

func run(_ *cobra.Command, _ []string) error {
	log.Error().
		Err(fmt.Errorf("please use one of the supported sub-command")).
		Send()
	return nil
}
