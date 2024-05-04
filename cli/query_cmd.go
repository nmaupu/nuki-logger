package cli

import (
	"fmt"
	"slices"
	"time"

	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/model"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	FlagLimit    = "limit"
	FlagFromDate = "from"
	FlagToDate   = "to"
	FlagJson     = "json"

	FromDateTime = "fromDateTime"
	ToDateTime   = "toDateTime"
)

type QueryConfig struct {
	Config
}

var (
	QueryCmd = &cobra.Command{
		Use:   "query",
		Short: "Query logs from the Nuki API",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			lim := viper.GetInt(FlagLimit)
			if lim <= 0 || lim > 50 {
				return fmt.Errorf("limit is out of bound. Should be between 1 and 50")
			}

			if viper.GetString(FlagFromDate) != "" {
				fromDate, err := time.Parse(time.RFC3339, viper.GetString(FlagFromDate))
				if err != nil {
					return err
				}
				viper.Set(FromDateTime, fromDate)
			}
			if viper.GetString(FlagToDate) != "" {
				toDate, err := time.Parse(time.RFC3339, viper.GetString(FlagToDate))
				if err != nil {
					return err
				}
				viper.Set(ToDateTime, toDate)
			}

			return nil
		},
		RunE: QueryRun,
	}
)

func init() {
	QueryCmd.Flags().IntP(FlagLimit, "l", 20, "Limits number of logs returned by the Nuki API (max: 50)")
	QueryCmd.Flags().String(FlagFromDate, "", "Retrieve logs from this date (RFC3339")
	QueryCmd.Flags().String(FlagToDate, "", "Retrieve logs to this date (RFC3339)")
	QueryCmd.Flags().Bool(FlagJson, false, "Output results in json")

	_ = viper.BindPFlags(QueryCmd.Flags())
}

func QueryRun(_ *cobra.Command, _ []string) error {
	logsReader := config.LogsReader
	logsReader.Limit = viper.GetInt(FlagLimit)
	logsReader.FromDate = viper.GetTime(FromDateTime)
	logsReader.ToDate = viper.GetTime(ToDateTime)
	logs, err := logsReader.Execute()
	if err != nil {
		return err
	}

	slices.Reverse(logs)

	var events []*messaging.Event
	for _, l := range logs {
		var reservationName string
		if l.Trigger == model.NukiTriggerKeypad && l.Source == model.NukiSourceKeypadCode && l.State != model.NukiStateWrongKeypadCode {
			reservationName, err = config.ReservationsReader.GetReservationName(l.Name)
			if err != nil {
				log.Error().
					Err(err).
					Str("ref", l.Name).
					Msg("Unable to get reservation's name, keeping original ref as name")
				reservationName = l.Name
			}
		}

		events = append(events, &messaging.Event{
			Log:             l,
			ReservationName: reservationName,
			Json:            viper.GetBool(FlagJson),
		})
	}

	for _, sender := range senders {
		if err := sender.Send(events); err != nil {
			log.Error().
				Err(err).
				Str("sender", sender.GetName()).
				Msg("Unable to send message")
		}
	}

	return nil
}
