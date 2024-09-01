package picker

import (
	"fmt"

	"github.com/eniehack/planet-eniehack/internal/model"
	"github.com/jmoiron/sqlx"
)

// RecipeScraper インターフェースを定義
type FeedPicker interface {
	Pick() error
}

type Source struct {
	Id        int    `db:"id"`
	SourceUrl string `db:"source_url"`
	SiteUrl   string `db:"site_url"`
	Type      int    `db:"type"`
}

func PickerFactory(db *sqlx.DB, src *Source) (FeedPicker, error) {
	switch src.Type {
	case model.TYPE_MASTODON:
		h := new(MastodonHandler)
		h.DB = db
		h.SiteConfig = src
		return h, nil
	case model.TYPE_MISSKEY:
		h := new(MisskeyHandler)
		h.DB = db
		h.SiteConfig = src
		return h, nil
	case model.TYPE_SCRAPBOX:
		h := new(ScrapboxHandler)
		h.DB = db
		h.SiteConfig = src
		return h, nil
	case model.TYPE_BLOG:
		h := new(BlogHandler)
		h.DB = db
		h.SiteConfig = src
		return h, nil
	default:
		return nil, fmt.Errorf("unsupported type site: %d", src.Id)
	}
}
