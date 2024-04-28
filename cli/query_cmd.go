package cli

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"nuki-logger/nukiapi"
	"time"
)

const (
	FlagLimit    = "limit"
	FlagFromDate = "from"
	FlagToDate   = "to"
	FromDateTime = "fromDateTime"
	ToDateTime   = "toDateTime"
)

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

	viper.BindPFlags(QueryCmd.Flags())
}

func QueryRun(cmd *cobra.Command, args []string) error {
	nukiLogsReader := nukiapi.LogsReader{
		SmartlockID: viper.GetString(PersistentFlagSmartlockID),
		Token:       viper.GetString(PersistentFlagToken),
		Limit:       viper.GetInt(FlagLimit),
		FromDate:    viper.GetTime(FromDateTime),
		ToDate:      viper.GetTime(ToDateTime),
	}
	logs, err := nukiLogsReader.Execute()
	if err != nil {
		return err
	}

	for _, l := range logs {
		log.Info().
			Time("date", l.Date).
			Str("source", l.Source.String()).
			Str("action", l.Action.String()).
			Str("state", l.State.String()).
			Str("trigger", l.Trigger.String()).
			Str("name", l.Name).
			Send()
	}

	return nil
}
