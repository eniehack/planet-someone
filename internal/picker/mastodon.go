package picker

import (
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/oklog/ulid/v2"
)

type MastodonHandler struct {
	BaseHandler
}

func (h *MastodonHandler) Pick() error {
	lastRun, err := h.ReadLastRunTime(h.SiteConfig.Id)
	if err != nil {
		fmt.Println("Error reading last run time:", err)
		lastRun = time.Time{}
	}
	feed, err := gofeed.NewParser().ParseURL(h.SiteConfig.SourceUrl)
	if err != nil {
		return fmt.Errorf("error parsing RSS feed: %s", err)
	}
	stmt, err := h.DB.Prepare("INSERT INTO posts (id, title, url, posts_source, date) VALUES (?, ?, ?, ?, ?);")
	if err != nil {
		return fmt.Errorf("cannot make prepare statement: %s", err)
	}
	// 新しい記事を探す
	for _, item := range feed.Items {
		if item.PublishedParsed.After(lastRun) {
			id := ulid.Make()
			if _, err := stmt.Exec(id.String(), item.Title, item.Link, h.SiteConfig.Id, item.PublishedParsed.Format(time.RFC3339)); err != nil {
				return fmt.Errorf("cannot insert item(%s): %s", item.Link, err)
			}
		}
	}
	// 現在の時刻を保存
	if err = h.SaveLastRunTime(time.Now(), h.SiteConfig.Id); err != nil {
		return fmt.Errorf("error saving last run time: %s", err)
	}
	return nil
}
