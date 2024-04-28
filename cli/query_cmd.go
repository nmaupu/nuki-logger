package cli

import (
	"fmt"
	"github.com/nmaupu/nuki-logger/messaging"
	"github.com/nmaupu/nuki-logger/nukiapi"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"time"
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

	viper.BindPFlags(QueryCmd.Flags())
}

func QueryRun(cmd *cobra.Command, args []string) error {
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
			if err := sender.Send(&messaging.Event{
				Log:  l,
				Json: viper.GetBool(FlagJson),
			}); err != nil {
				log.Error().
					Err(err).
					Str("sender", sender.GetName()).
					Msg("Unable to send message to sender")
			}
		}
	}

	return nil
}
