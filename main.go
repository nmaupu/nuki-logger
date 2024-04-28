package main

import (
	"github.com/nmaupu/nuki-logger/cli"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true, TimeFormat: time.RFC3339})

	if err := cli.RootCmd.Execute(); err != nil {
		log.Error().
			Err(err).
			Send()
	}
}
