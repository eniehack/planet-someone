package picker

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type BaseHandler struct {
	DB         *sqlx.DB
	SiteConfig *Source
}

func (h *BaseHandler) ReadLastRunTime(src int) (time.Time, error) {
	row := h.DB.QueryRow("SELECT timestamp FROM crawl_time WHERE source = ? ORDER BY timestamp ASC;", src)
	if row.Err() != nil {
		return time.Time{}, row.Err()
	}
	var timestamp_unit int64
	row.Scan(&timestamp_unit)
	timestamp := time.Unix(timestamp_unit, 0)
	return timestamp, nil
}

func (h *BaseHandler) SaveLastRunTime(t time.Time, src int) error {
	if _, err := h.DB.Exec("INSERT INTO crawl_time VALUES (?, ?);", t.Unix(), src); err != nil {
		return err
	}
	return nil
}
