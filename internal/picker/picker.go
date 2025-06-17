package picker

import (
	"fmt"
	"time"

	"github.com/eniehack/planet-someone/internal/config"
	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
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

func PickerFactory(db *sqlx.DB, src config.SiteConfig) (FeedPicker, error) {
	switch config := src.(type) {
	case *config.MastodonConfig:
		h := new(MastodonHandler)
		h.DB = db
		h.Config = config
		return h, nil
	case *config.ScrapboxConfig:
		h := new(ScrapboxHandler)
		h.DB = db
		h.Config = config
		return h, nil
	case *config.BlogConfig:
		h := new(BlogHandler)
		h.DB = db
		h.Config = config
		return h, nil
	case *config.MisskeyConfig:
		h := new(MisskeyHandler)
		h.DB = db
		h.Config = config
		return h, nil
	default:
		return nil, fmt.Errorf("unsupported type site: %s", config.GetType())
	}
}

func BuildID(t *time.Time) string {
	id := ulid.Make()
	id.SetTime(uint64(t.Unix()))
	return id.String()
}
