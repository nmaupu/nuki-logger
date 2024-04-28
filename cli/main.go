package cli

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"nuki-logger/cache"
	"nuki-logger/model"
	"nuki-logger/nukiapi"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	EnvVarsPrefix    = "NUKI_LOGGER"
	FlagToken        = "token"
	FlagSmartlockID  = "smartlockid"
	FlagInterval     = "interval"
	IntervalDuration = "intervalDuration"
)

var (
	requiredFlags = []string{
		FlagToken,
		FlagSmartlockID,
	}
	NukiLoggerCmd = &cobra.Command{
		Use:           "log",
		Short:         "Start a process that listens for Nuki logs and forwards them to a messaging service",
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

			interval := viper.GetString(FlagInterval)
			dur, err := time.ParseDuration(interval)
			if err != nil {
				return fmt.Errorf("interval is not a correct duration")
			}
			viper.Set(IntervalDuration, dur)
			return nil
		},
		RunE: run,
	}
)

func init() {
	NukiLoggerCmd.Flags().StringP(FlagSmartlockID, "s", "", "Smartlock ID to get the logs from")
	NukiLoggerCmd.Flags().StringP(FlagToken, "t", "", "Token to access the Nuki API")
	NukiLoggerCmd.Flags().StringP(FlagInterval, "i", "10s", "Interval used to retrieve logs")

	viper.AutomaticEnv()
	viper.SetEnvPrefix(EnvVarsPrefix)
	_ = viper.BindPFlags(NukiLoggerCmd.Flags())
}

func run(cmd *cobra.Command, args []string) error {
	ticker := time.NewTicker(viper.GetDuration(IntervalDuration))
	inter := make(chan os.Signal, 1)
	signal.Notify(inter, syscall.SIGINT, syscall.SIGTERM)
	smartlockID := viper.GetString(FlagSmartlockID)
	token := viper.GetString(FlagToken)

	log.Info().Msg("Reading old log responses from cache")
	cacheLogs, err := cache.LoadCacheFromDisk()
	if err != nil {
		return err
	}
	if len(cacheLogs) == 0 {
		// No cache, creating one
		log.Info().Msg("No cache yet, creating one")
		cacheLogs, err := nukiapi.ReadLogs(smartlockID, token)
		if err != nil {
			return err
		}
		cache.SaveCacheToDisk(cacheLogs)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Info().Msg("Getting logs from api")
				newResponses, err := nukiapi.ReadLogs(smartlockID, token)
				if err != nil {
					log.Error().Err(err).Msg("An error occurred getting logs from API")
				}

				diff := model.Diff(newResponses, cacheLogs)
				if len(diff) > 0 {
					for _, d := range diff {
						// log those new messages
						log.Info().
							Time("date", d.Date).
							Str("source", d.Source.String()).
							Str("action", d.Action.String()).
							Str("state", d.State.String()).
							Str("trigger", d.Trigger.String()).
							Str("name", d.Name).
							Msg("New log")
					}

					cacheLogs = newResponses
					cache.SaveCacheToDisk(cacheLogs)
				}
			case <-inter:
				log.Info().Msg("Stopping.")
				ticker.Stop()
				wg.Done()
			}
		}
	}()
	wg.Wait()
	return nil
}
