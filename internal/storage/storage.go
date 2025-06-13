package storage

import (
	"context"
	"github.com/moorzeen/banner-stat/internal/model"
	"time"
)

type Storage interface {
	IncrementClicks(ctx context.Context, bannerID int) error
	GetStats(ctx context.Context, bannerID int, from, to time.Time) ([]model.ClickStats, error)
	Close() error
}
