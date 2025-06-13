package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/moorzeen/banner-stat/internal/model"
	"github.com/moorzeen/banner-stat/internal/storage"
	"github.com/rs/zerolog/log"
	"strconv"
	"time"
)

type API struct {
	http    *fiber.App
	storage storage.Storage
}

func New(storage storage.Storage) *API {
	a := &API{
		http: fiber.New(fiber.Config{
			ServerHeader: "",
			WriteTimeout: 5 * time.Second,
			ReadTimeout:  5 * time.Second,
			IdleTimeout:  10 * time.Second,
			BodyLimit:    4 * 1024,
		}),

		storage: storage,
	}

	a.http.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST",
		AllowHeaders: "Content-Type",
	}))

	a.http.Use(recover.New())
	a.http.Use(logger.New())

	//a.http.Use(limiter.New(limiter.Config{
	//	Max:        1000,
	//	Expiration: time.Second,
	//}))

	a.setupRoutes()

	return a
}

func (a *API) setupRoutes() {
	a.http.Post("/counter/:id", a.IncrementClicks)
	a.http.Post("/stats/:id", a.GetStats)
}

func (a *API) IncrementClicks(c *fiber.Ctx) error {
	bannerID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		log.Err(err).Msg("invalid banner id")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid banner ID",
		})
	}

	if err := a.storage.IncrementClicks(c.Context(), bannerID); err != nil {
		log.Err(err).Msg("internal api error")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal api error",
		})
	}

	return c.SendStatus(fiber.StatusOK)
}

func (a *API) GetStats(c *fiber.Ctx) error {
	bannerID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		log.Err(err).Msg("invalid banner id")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid banner ID",
		})
	}

	var req model.StatsRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid request body", err)
	}

	from := time.Time(req.From)
	to := time.Time(req.To)

	if from.After(to) {
		return badRequest(c, "`from` must be before `to`", nil)
	}

	stats, err := a.storage.GetStats(c.Context(), bannerID, from, to)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal api error",
		})
	}

	return c.JSON(model.StatsResponse{Stats: stats})
}

func (a *API) Listen(addr string) error {
	return a.http.Listen(addr)
}

func (a *API) Shutdown() error {
	if err := a.http.Shutdown(); err != nil {
		return err
	}

	if err := a.storage.Close(); err != nil {
		return err
	}

	return nil
}

func badRequest(c *fiber.Ctx, msg string, err error) error {
	if err != nil {
		log.Warn().Err(err).Msg(msg)
	}
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
}
