package picker

import (
	"database/sql"
	"time"

	"github.com/eniehack/planet-someone/internal/config"
	"github.com/jmoiron/sqlx"
)

var DEFAULT_DURATION time.Duration = time.Hour * 24 * 14

type BaseHandler struct {
	DB         *sqlx.DB
	SiteConfig *config.SiteConfig
}

func (h *BaseHandler) ReadLastRunTime(dur *time.Duration) (*time.Time, error) {
	row := h.DB.QueryRow("SELECT created_at FROM posts WHERE src = ? ORDER BY created_at DESC;", h.SiteConfig.Id)
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			t := time.Now().Add(*dur)
			return &t, row.Err()
		}
		return nil, row.Err()
	}
	var timestamp_unit int64
	row.Scan(&timestamp_unit)
	timestamp := time.Unix(timestamp_unit, 0)
	return &timestamp, nil
}

func (h *BaseHandler) SaveLastRunTime(t time.Time, src int) error {
	if _, err := h.DB.Exec("INSERT OR REPLACE INTO crawl_time (timestamp, source) VALUES (?, ?);", t.Unix(), src); err != nil {
		return err
	}
	return nil
}
