package cli

import (
	"fmt"
	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

const (
	PersistentFlagConfig = "config"
	PersistentFlagSender = "sender"
)

var (
	requiredStringFlags = []string{
		PersistentFlagConfig,
	}
	requiredStringSliceFlags = []string{
		PersistentFlagSender,
	}

	config  = Config{}
	senders []messaging.Sender

	RootCmd = &cobra.Command{
		Use:           "nuki-logger",
		Short:         "Query Nuki api for most recent logs",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if strings.Contains(cmd.CommandPath(), "version") {
				return nil
			}

			var requiredFlagsMissing []string
			for _, v := range requiredStringFlags {
				if viper.GetString(v) == "" {
					requiredFlagsMissing = append(requiredFlagsMissing, v)
				}
			}
			for _, v := range requiredStringSliceFlags {
				if len(viper.GetStringSlice(v)) == 0 {
					requiredFlagsMissing = append(requiredFlagsMissing, v)
				}
			}

			if len(requiredFlagsMissing) > 0 {
				return fmt.Errorf("the following flag(s) are required: %s", strings.Join(requiredFlagsMissing, ", "))
			}

			if viper.GetString(PersistentFlagConfig) != "" {
				viper.SetConfigName(viper.GetString(PersistentFlagConfig))
			}

			if err := config.LoadConfig(viper.GetViper()); err != nil {
				return err
			}

			return initSenders()
		},
		RunE: run,
	}
)

func init() {
	RootCmd.PersistentFlags().StringP(PersistentFlagConfig, "c", "", "Configuration file")
	RootCmd.PersistentFlags().StringSliceP(PersistentFlagSender, "s", []string{}, "Senders to send new logs to")

	RootCmd.AddCommand(VersionCmd)
	RootCmd.AddCommand(QueryCmd)
	RootCmd.AddCommand(ServerCmd)

	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	_ = viper.BindPFlags(RootCmd.PersistentFlags())
}

func initSenders() error {
	for _, v := range viper.GetStringSlice(PersistentFlagSender) {
		s, err := config.GetSender(v)
		if err != nil {
			log.Error().Err(err).Msgf("unable to use sender %s", v)
			continue
		}
		senders = append(senders, s)
	}

	if len(senders) == 0 {
		return fmt.Errorf("no sender available, aborting")
	}
	return nil
}

func run(_ *cobra.Command, _ []string) error {
	log.Debug().
		Msg(fmt.Sprintf("%+v", config))
	log.Error().
		Err(fmt.Errorf("please use one of the supported sub-command")).
		Send()
	return nil
}
