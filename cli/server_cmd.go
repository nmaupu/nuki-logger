package cli

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/nmaupu/nuki-logger/cache"
	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/model"
	"github.com/nmaupu/nuki-logger/nukiapi"
	"github.com/nmaupu/nuki-logger/telegrambot"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
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

	logsReader      nukiapi.LogsReader
	smartlockReader nukiapi.SmartlockReader
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

	logsReader = nukiapi.LogsReader{
		APICaller:   nukiapi.APICaller{Token: config.NukiAPIToken},
		SmartlockID: config.SmartlockID,
		Limit:       20,
	}
	smartlockReader = nukiapi.SmartlockReader{
		APICaller:   nukiapi.APICaller{Token: config.NukiAPIToken},
		SmartlockID: config.SmartlockID,
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

	if config.TelegramBot.Enabled {
		if err := runTelegramBot(); err != nil {
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
				resp, err := smartlockReader.Execute()
				if err != nil {
					log.Error().Err(err).Msg("Unable to check smartlock")
				}
				if resp.State.BatteryCritical ||
					resp.State.KeypadBatteryCritical ||
					resp.State.DoorsensorBatteryCritical ||
					resp.State.BatteryCharge <= 30 {
					for _, sender := range senders {
						if err := sender.Send(&messaging.Event{
							Smartlock: *resp,
						}); err != nil {
							log.Error().
								Err(err).
								Str("sender", sender.GetName()).
								Msg("Unable to send message to sender")
						}
					}
				}

			case <-tickerLogs.C:
				log.Info().Msg("Getting logs from api")

				newResponses, err := logsReader.Execute()
				if err != nil {
					log.Error().Err(err).Msg("An error occurred getting logs from API")
				}

				diff := model.Diff(newResponses, cacheLogs)
				if len(diff) > 0 {
					for _, d := range diff {
						var reservationName string
						if d.Trigger == model.NukiTriggerKeypad && d.Source == model.NukiSourceKeypadCode && d.State != model.NukiStateWrongKeypadCode {
							reservationName, err = getReservationName(d.Name, &config)
							if err != nil {
								log.Error().
									Err(err).
									Str("ref", d.Name).
									Msg("Unable to get reservation's name, keeping original ref as name")
								reservationName = d.Name
							}
						}

						// log those new messages
						for _, sender := range senders {
							if err := sender.Send(&messaging.Event{
								Log:             d,
								ReservationName: reservationName,
							}); err != nil {
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
				tickerLogs.Stop()
				tickerSmartlock.Stop()
				wg.Done()
			}
		}
	}()

	wg.Wait()
	return nil
}

func runTelegramBot() error {
	tgSenderInterface, err := config.GetSender(config.TelegramBot.SenderName)
	if err != nil {
		return err
	}
	tgSender := tgSenderInterface.(*messaging.TelegramSender)

	commandNames := []string{
		"/help",
		"/battery",
		"/code",
	}
	commands := telegrambot.Commands{}
	commands["help"] = func(msg *tgbotapi.MessageConfig) {
		msg.Text = fmt.Sprintf("The following commands are available: %s", strings.Join(commandNames, ", "))
	}
	fBattery := func(msg *tgbotapi.MessageConfig) {
		res, err := smartlockReader.Execute()
		if err != nil {
			msg.Text = fmt.Sprintf("Unable to read smartlock status from API, err=%v", err)
		} else {
			msg.Text = res.PrettyFormat()
		}
	}
	commands["battery"] = fBattery
	commands["bat"] = fBattery
	commands["code"] = func(msg *tgbotapi.MessageConfig) {
		msg.Text = "Not implemented yet."
	}

	return commands.Start(tgSender)
}
