package main

import (
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/beeper/alertmanager-to-bigquery/internal"
	"github.com/beeper/alertmanager-to-bigquery/internal/config"
)

func main() {
	configFilename := flag.String("config", "config.yaml", "Config filename")
	prettyLogs := flag.Bool("prettyLogs", false, "Display pretty logs")
	debug := flag.Bool("debug", false, "Enable debug logging")

	flag.Parse()

	if *prettyLogs {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug logging enabled")
	}

	log.Debug().Msgf("Loading config file: %s", *configFilename)
	cfg := config.LoadConfigFile(*configFilename)

	amtobq := internal.NewAlertManagerToBigQuery(cfg)
	amtobq.Start()
}
