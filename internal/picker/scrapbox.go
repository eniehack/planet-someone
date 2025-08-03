package picker

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/eniehack/planet-someone/internal/config"
	"github.com/mmcdole/gofeed"
)

type ScrapboxHandler struct {
	BaseHandler
	Config *config.ScrapboxConfig
}

func (h *ScrapboxHandler) Pick() error {
	lastRun, err := h.ReadLastRunTime(&DEFAULT_DURATION)
	if err != nil {
		slog.Info(fmt.Sprintf("Error reading last run time: %s", err))
	}
	feed, err := gofeed.NewParser().ParseURL(h.Config.SourceUrl)
	if err != nil {
		return fmt.Errorf("error parsing rss feed: %s", err)
	}
	stmt, err := h.DB.Prepare("INSERT INTO posts (id, content, url, src, type, created_at) VALUES (?, ?, ?, ?, ?, ?);")
	if err != nil {
		return fmt.Errorf("cannot make prepare statement: %s", err)
	}
	// 新しい記事を探す
	for _, item := range feed.Items {
		if lastRun.Unix() < item.PublishedParsed.Unix() {
			id := BuildID(item.PublishedParsed)
			if _, err := stmt.Exec(id, item.Title, item.Link, h.Config.Id, h.Config.Type, item.PublishedParsed.Unix()); err != nil {
				return fmt.Errorf("cannot insert item(%s): %s", item.Link, err)
			}
		}
	}
	return nil
}

func (h *ScrapboxHandler) ReadLastRunTime(dur *time.Duration) (*time.Time, error) {
	row := h.DB.QueryRow("SELECT created_at FROM posts WHERE src = ? ORDER BY created_at DESC;", h.Config.Id)
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
