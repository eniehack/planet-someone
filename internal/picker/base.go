package picker

import (
	"time"

	"github.com/jmoiron/sqlx"
)

var DEFAULT_DURATION time.Duration = time.Hour * 24 * 14

type BaseHandler struct {
	DB         *sqlx.DB
	SiteConfig *Source
}

func (h *BaseHandler) ReadLastRunTime(src int, dur *time.Duration) (*time.Time, error) {
	row := h.DB.QueryRow("SELECT unixepoch(date) FROM posts WHERE src = ? ORDER BY date DESC;", src)
	if row.Err() != nil {
		t := time.Now().Add(*dur)
		return &t, row.Err()
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
