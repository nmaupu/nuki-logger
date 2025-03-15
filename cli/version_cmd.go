package cli

import (
	"github.com/nmaupu/nuki-logger/model"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	VersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print app version",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info().
				Str("app_name", model.AppName).
				Str("version", model.ApplicationVersion).
				Str("build_date", model.BuildDate).
				Send()
		},
	}
)
