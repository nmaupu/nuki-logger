package cli

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"nuki-logger/cache"
	"nuki-logger/messaging"
	"nuki-logger/model"
	"nuki-logger/nukiapi"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	FlagServerInterval = "interval"
)

var (
	ServerCmd = &cobra.Command{
		Use:           "server",
		Short:         "Run a server querying Nuki api for logs on a regular interval",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          RunServer,
	}
)

func init() {
	ServerCmd.Flags().DurationP(FlagServerInterval, "i", time.Second*60, "Interval at which to check new logs")
	_ = viper.BindPFlags(ServerCmd.Flags())
}

func RunServer(_ *cobra.Command, _ []string) error {
	log.Debug().Dur(FlagServerInterval, viper.GetDuration(FlagServerInterval)).Send()
	ticker := time.NewTicker(viper.GetDuration(FlagServerInterval))
	interruptSigChan := make(chan os.Signal, 1)
	signal.Notify(interruptSigChan, syscall.SIGINT, syscall.SIGTERM)

	logsReader := nukiapi.LogsReader{
		SmartlockID: config.SmartlockID,
		Token:       config.NukiAPIToken,
		Limit:       20,
	}

	log.Info().Msg("Reading old log responses from cache")
	cacheLogs, err := cache.LoadCacheFromDisk()
	if err != nil {
		return err
	}
	if len(cacheLogs) == 0 {
		// No cache, creating one
		log.Info().Msg("No cache yet, creating one")
		cacheLogs, err := logsReader.Execute()
		if err != nil {
			return err
		}
		if err := cache.SaveCacheToDisk(cacheLogs); err != nil {
			return err
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Info().Msg("Getting logs from api")

				newResponses, err := logsReader.Execute()
				if err != nil {
					log.Error().Err(err).Msg("An error occurred getting logs from API")
				}

				diff := model.Diff(newResponses, cacheLogs)
				if len(diff) > 0 {
					for _, d := range diff {
						// log those new messages
						for _, sender := range senders {
							if err := sender.Send(&messaging.Event{Log: d}); err != nil {
								log.Error().
									Err(err).
									Str("sender", sender.GetName()).
									Msg("Unable to send message to sender")
							}
						}
					}

					cacheLogs = newResponses
					if err := cache.SaveCacheToDisk(cacheLogs); err != nil {
						log.Error().Err(err).Msg("Unable to save cache file to disk")
					}
				}
			case <-interruptSigChan:
				log.Info().Msg("Stopping.")
				ticker.Stop()
				wg.Done()
			}
		}
	}()
	wg.Wait()
	return nil
}
