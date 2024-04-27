package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"nuki-logger/nukiapi"
	"strings"
)

const (
	EnvVarsPrefix   = "NUKI_LOGGER"
	FlagToken       = "token"
	FlagSmartlockID = "smartlockid"
)

var (
	NukiLoggerCmd = &cobra.Command{
		Use:           "log",
		Short:         "Start a process that listens for Nuki logs and forwards them to a messaging service",
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var requiredFlags []string
			if viper.GetString(FlagToken) == "" {
				requiredFlags = append(requiredFlags, FlagToken)
			}
			if viper.GetString(FlagSmartlockID) == "" {
				requiredFlags = append(requiredFlags, FlagSmartlockID)
			}
			if len(requiredFlags) > 0 {
				return fmt.Errorf("the following flag(s) are required: %s", strings.Join(requiredFlags, ", "))
			}
			return nil
		},
		RunE: run,
	}
)

func init() {
	NukiLoggerCmd.Flags().StringP(FlagSmartlockID, "s", "", "Smartlock ID to get the logs from")
	NukiLoggerCmd.Flags().StringP(FlagToken, "t", "", "Token to access the Nuki API")

	viper.AutomaticEnv()
	viper.SetEnvPrefix(EnvVarsPrefix)
	_ = viper.BindPFlags(NukiLoggerCmd.Flags())
}

func run(cmd *cobra.Command, args []string) error {
	return nukiapi.ReadLog(viper.GetString(FlagSmartlockID), viper.GetString(FlagToken))
}
