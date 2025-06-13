package app

import (
	"fmt"
	"github.com/moorzeen/banner-stat/internal/api"
	"github.com/moorzeen/banner-stat/internal/app/config"
	"github.com/moorzeen/banner-stat/internal/storage/postgres"
	"github.com/rs/zerolog/log"
)

type App interface {
	Run() error
	Stop() error
}

type app struct {
	api *api.API
	cfg *config.Config
}

func New(cfg *config.Config) (App, error) {
	st, err := postgres.NewStorage(cfg.DB)
	if err != nil {
		return nil, err
	}

	return &app{
		api: api.New(st),
		cfg: cfg,
	}, nil
}

func (a *app) Run() error {
	err := a.api.Listen(fmt.Sprintf(":%s", a.cfg.Port))
	if err != nil {
		return err
	}

	return nil
}

func (a *app) Stop() error {
	err := a.api.Shutdown()
	if err != nil {
		log.Fatal().Err(err).Msg("api shutdown error")
	}

	return nil
}
