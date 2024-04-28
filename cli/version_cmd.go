package cli

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	BuildDate          string
	ApplicationVersion string
)

const (
	AppName = "nuki-logger"
)

var (
	VersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print app version",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info().
				Str("app_name", AppName).
				Str("version", ApplicationVersion).
				Str("build_date", BuildDate).
				Send()
		},
	}
)
