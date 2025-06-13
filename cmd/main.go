package main

import (
	"github.com/moorzeen/banner-stat/internal/app"
	"github.com/moorzeen/banner-stat/internal/app/config"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.DateTime,
	}).With().Caller().Logger()

	log.Debug().Msg("debug mode enabled")

	application, err := app.New(config.GetConfig())
	if err != nil {
		log.Fatal().Err(err).Msg("application initialization error")
	}

	go func() {
		if err := application.Run(); err != nil {
			log.Fatal().Err(err).Msg("application running error")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	log.Info().Msg("application started")

	<-stop

	log.Info().Msg("shutting down...")

	err = application.Stop()
	if err != nil {
		log.Fatal().Err(err).Msg("application stop error")
	}

	log.Info().Msg("stopped")
}
