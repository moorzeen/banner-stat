package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/moorzeen/banner-stat/internal/model"
	"github.com/moorzeen/banner-stat/internal/storage"
	"github.com/rs/zerolog/log"
	"time"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func WaitForDatabase(db *sql.DB, maxAttempts int, delay time.Duration) error {
	for i := 0; i < maxAttempts; i++ {
		err := db.PingContext(context.Background())
		if err == nil {
			return nil
		}

		log.Warn().Err(err).
			Int("attempt", i+1).
			Int("maxAttempts", maxAttempts).
			Msg("failed to connect to the database, retrying...")

		time.Sleep(delay)
	}

	return fmt.Errorf("failed to connect to the database after %d attempts", maxAttempts)
}

func NewStorage(dbURL string) (storage.Storage, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)

	if err := WaitForDatabase(db, 30, time.Second); err != nil {
		return nil, fmt.Errorf("database is not ready: %w", err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS banner_clicks (
            banner_id INTEGER NOT NULL,
            click_time TIMESTAMP NOT NULL,
            minute_bucket TIMESTAMP NOT NULL,
            clicks INTEGER DEFAULT 1,
            CONSTRAINT banner_clicks_unique UNIQUE (banner_id, minute_bucket)
        );
        
        CREATE INDEX IF NOT EXISTS idx_banner_clicks_minute 
        ON banner_clicks(banner_id, minute_bucket);

    `)

	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (s *Storage) IncrementClicks(ctx context.Context, bannerID int) error {
	now := time.Now()
	minuteBucket := now.Truncate(time.Minute)

	_, err := s.db.ExecContext(ctx, `
        INSERT INTO banner_clicks (banner_id, click_time, minute_bucket, clicks)
        VALUES ($1, $2, $3, 1)
        ON CONFLICT (banner_id, minute_bucket)
        DO UPDATE SET clicks = banner_clicks.clicks + 1
    `, bannerID, now, minuteBucket)

	return err
}

func (s *Storage) GetStats(ctx context.Context, bannerID int, from, to time.Time) ([]model.ClickStats, error) {
	rows, err := s.db.QueryContext(ctx, `
        SELECT minute_bucket, clicks
        FROM banner_clicks
        WHERE banner_id = $1 
        AND minute_bucket >= $2 
        AND minute_bucket <= $3
        ORDER BY minute_bucket
    `, bannerID, from, to)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]model.ClickStats, 0)
	for rows.Next() {
		var stat model.ClickStats
		if err := rows.Scan(&stat.Timestamp, &stat.Value); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (s *Storage) Close() error {
	if err := s.db.Close(); err != nil {
		return err
	}

	return nil
}
