package picker

import (
	"time"

	"github.com/jmoiron/sqlx"
)

var DEFAULT_DURATION time.Duration = time.Hour * 24 * 14

type BaseHandler struct {
	DB *sqlx.DB
}

type Handler interface {
	ReadLastRunTime(dur *time.Duration) (*time.Time, error)
}

func (h *BaseHandler) SaveLastRunTime(t time.Time, src int) error {
	if _, err := h.DB.Exec("INSERT OR REPLACE INTO crawl_time (timestamp, source) VALUES (?, ?);", t.Unix(), src); err != nil {
		return err
	}
	return nil
}
