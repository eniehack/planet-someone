package picker

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/mmcdole/gofeed"
)

type ScrapboxHandler struct {
	BaseHandler
}

func (h *ScrapboxHandler) Pick() error {
	lastRun, err := h.ReadLastRunTime(&DEFAULT_DURATION)
	if err != nil {
		slog.Info(fmt.Sprintf("Error reading last run time: %s", err))
	}
	feed, err := gofeed.NewParser().ParseURL(h.SiteConfig.SourceUrl)
	if err != nil {
		return fmt.Errorf("error parsing rss feed: %s", err)
	}
	stmt, err := h.DB.Prepare("INSERT INTO posts (id, title, url, src, created_at) VALUES (?, ?, ?, ?, ?);")
	if err != nil {
		return fmt.Errorf("cannot make prepare statement: %s", err)
	}
	// 新しい記事を探す
	for _, item := range feed.Items {
		fmt.Printf("lastrun: %s, published: %s\n", lastRun.Format(time.RFC3339), item.PublishedParsed.Format(time.RFC3339))
		if lastRun.Unix() < item.PublishedParsed.Unix() {
			id := BuildID(item.PublishedParsed)
			if _, err := stmt.Exec(id, item.Title, item.Link, h.SiteConfig.Id, item.PublishedParsed.Unix()); err != nil {
				return fmt.Errorf("cannot insert item(%s): %s", item.Link, err)
			}
		}
	}
	return nil
}
