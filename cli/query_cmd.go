package cli

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"nuki-logger/messaging"
	"nuki-logger/nukiapi"
	"time"
)

const (
	FlagLimit    = "limit"
	FlagFromDate = "from"
	FlagToDate   = "to"
	FlagSender   = "sender"

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

			if len(viper.GetStringSlice(FlagSender)) == 0 {
				return fmt.Errorf("specify at least one sender to send log to")
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
	QueryCmd.Flags().StringSlice(FlagSender, []string{}, "Send results to this specific sender")

	viper.BindPFlags(QueryCmd.Flags())
}

func QueryRun(cmd *cobra.Command, args []string) error {
	var senders []messaging.Sender
	sendersFlag := viper.GetStringSlice(FlagSender)
	for _, v := range sendersFlag {
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

	nukiLogsReader := nukiapi.LogsReader{
		SmartlockID: config.SmartlockID,
		Token:       config.NukiAPIToken,
		Limit:       viper.GetInt(FlagLimit),
		FromDate:    viper.GetTime(FromDateTime),
		ToDate:      viper.GetTime(ToDateTime),
	}
	logs, err := nukiLogsReader.Execute()
	if err != nil {
		return err
	}

	for _, l := range logs {
		for _, sender := range senders {
			if err := sender.Send(&messaging.Event{Log: l}); err != nil {
				log.Error().
					Err(err).
					Msgf("Unable to send message to sender")
			}
		}
	}

	return nil
}
