package cli

import (
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/mymmrac/telego"
	"github.com/nmaupu/nuki-logger/cache"
	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/model"
	"github.com/nmaupu/nuki-logger/telegrambot"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	FlagServerInterval     = "interval"
	SmartlockCheckInterval = time.Hour * 2
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
	tickerLogs := time.NewTicker(viper.GetDuration(FlagServerInterval))
	tickerSmartlock := time.NewTicker(SmartlockCheckInterval)
	interruptSigChan := make(chan os.Signal, 1)
	signal.Notify(interruptSigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Msg("Reading old log responses from cache")
	cacheLogs, err := cache.LoadCacheFromDisk()
	if err != nil {
		return err
	}
	if len(cacheLogs) == 0 {
		// No cache, creating one
		log.Info().Msg("No cache yet, creating one")
		cacheLogs, err := config.LogsReader.Execute()
		if err != nil {
			return err
		}
		if err := cache.SaveCacheToDisk(cacheLogs); err != nil {
			return err
		}
	}

	if config.TelegramBot.Enabled {
		tgSenderInterface, err := config.GetSender(config.TelegramBot.SenderName)
		if err != nil {
			return err
		}
		tgSender := tgSenderInterface.(*messaging.TelegramSender)
		nukiBot := telegrambot.NewNukiBot(tgSender,
			config.LogsReader,
			config.SmartlockReader,
			config.ReservationsReader,
			config.SmartlockAuthReader)
		if len(config.TelegramBot.RestrictToChatIDs) > 0 {
			log.Info().
				Ints64("chat_ids", config.TelegramBot.RestrictToChatIDs).
				Msg("Restricting bot access")
			nukiBot.AddFilter(func(update telego.Update) bool {
				if update.Message == nil || !telegrambot.IsPrivateMessage(update) {
					return true
				}
				return slices.Contains(config.TelegramBot.RestrictToChatIDs, update.Message.From.ID)
			})
		}

		if err := nukiBot.Start(); err != nil {
			return err
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			select {
			case <-tickerSmartlock.C:
				log.Info().Msg("Checking smartlock for issues")
				resp, err := config.SmartlockReader.Execute()
				if err != nil {
					log.Error().Err(err).Msg("Unable to check smartlock")
				}
				if resp.State.BatteryCritical ||
					resp.State.KeypadBatteryCritical ||
					resp.State.DoorsensorBatteryCritical ||
					resp.State.BatteryCharge <= 30 {
					for _, sender := range senders {
						e := &messaging.Event{Smartlock: *resp}
						if err := sender.Send([]*messaging.Event{e}); err != nil {
							log.Error().
								Err(err).
								Str("sender", sender.GetName()).
								Msg("Unable to send message to sender")
						}
					}
				}

			case <-tickerLogs.C:
				log.Info().Msg("Getting logs from api")

				newResponses, err := config.LogsReader.Execute()
				if err != nil {
					log.Error().Err(err).Msg("An error occurred getting logs from API")
				}

				diff := model.Diff(newResponses, cacheLogs)
				var events []*messaging.Event
				if len(diff) > 0 {
					for _, d := range diff {
						reservationName := d.Name
						if d.Trigger == model.NukiTriggerKeypad && d.Source == model.NukiSourceKeypadCode && d.State != model.NukiStateWrongKeypadCode {
							reservationName, err = config.ReservationsReader.GetReservationName(d.Name)
							if err != nil {
								log.Error().
									Err(err).
									Str("ref", d.Name).
									Msg("Unable to get reservation's name, keeping original ref as name")
								reservationName = d.Name
							}
						}

						events = append(events, &messaging.Event{
							Log:             d,
							ReservationName: reservationName,
						})
					}

					// log those new messages
					for _, sender := range senders {
						if err := sender.Send(events); err != nil {
							log.Error().
								Err(err).
								Str("sender", sender.GetName()).
								Msg("Unable to send message")
						}
					}

					cacheLogs = newResponses
					if err := cache.SaveCacheToDisk(cacheLogs); err != nil {
						log.Error().Err(err).Msg("Unable to save cache file to disk")
					}
				}
			case <-interruptSigChan:
				log.Info().Msg("Stopping.")
				tickerLogs.Stop()
				tickerSmartlock.Stop()
				wg.Done()
			}
		}
	}()

	wg.Wait()
	return nil
}
