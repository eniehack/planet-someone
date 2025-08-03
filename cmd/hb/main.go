package main

import (
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path"
	"sort"
	"time"

	_ "time/tzdata"

	"github.com/eniehack/planet-someone/internal/config"
	"github.com/eniehack/planet-someone/internal/hb"
	"github.com/eniehack/planet-someone/internal/model"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var GitRevision = "unknown"

func main() {
	var configFilePath string
	flag.StringVar(&configFilePath, "config", "./config.yml", "config file")
	flag.Parse()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	c := config.ReadConfig(configFilePath)
	db, err := sqlx.Connect("sqlite", fmt.Sprintf("file:%s", c.DB.DB))
	if err != nil {
		slog.Error(fmt.Sprintf("cannot open db: %s", err))
		os.Exit(1)
	}
	defer db.Close()
	tmpl, err := template.ParseFiles(path.Join(c.Hb.TemplateDir, "/index.html"))
	if err != nil {
		slog.Error(fmt.Sprintf("cannot parse template: %s", err))
		os.Exit(1)
	}
	hbConfig := new(hb.Config)
	hbConfig.Meta = hb.PageMeta{
		Url:         c.Hb.Url,
		Title:       c.Hb.Meta.Title,
		Description: c.Hb.Meta.Description,
	}
	posts := make(map[string][]hb.Post)
	tz, err := time.LoadLocation(c.Hb.TimeZone)
	if err != nil {
		slog.Error(fmt.Sprintf("cannot parse timezone: %s", err))
		os.Exit(1)
	}
	today := time.Now().In(tz)
	for i := today; today.Sub(i).Abs().Hours() <= (time.Hour * 24 * 14).Hours(); i = i.Add(time.Hour * -24) {
		dateStr := i.Format("2006-01-02")
		res, err := db.Query(
			`SELECT id, content, url, created_at, src
			 FROM posts
			 WHERE date(created_at, "unixepoch", "localtime") = ?
			 ORDER BY created_at DESC;`,
			dateStr,
		)
		if err != nil {
			slog.Error(fmt.Sprintf("cannot exec query: %s", err))
			os.Exit(1)
		}
		for res.Next() {
			post := hb.Post{}
			if err := res.Scan(
				&post.Id,
				&post.Content,
				&post.Url,
				&post.Date,
				&post.Src,
			); err != nil {
				slog.Error(fmt.Sprintf("cannot bind variable from query: %s", err))
				os.Exit(1)
			}
			tzAppliedDate := time.Unix(post.Date, 0).In(tz)
			post.ParsedDate = &tzAppliedDate
			posts[dateStr] = append(posts[dateStr], post)
		}
	}
	keys := make([]string, 0, len(posts))
	for k := range posts {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	sites := map[string]hb.Site{}
	for _, site := range c.Picker.Sites {
		switch site.Type {
		case model.TYPE_MASTODON:
			param, ok := site.RawParams.(*config.MastodonConfig)
			if !ok {
				return
			}
			sites[site.Id] = *param.GetMetadata()
		case model.TYPE_MISSKEY:
			param, ok := site.RawParams.(*config.MisskeyConfig)
			if !ok {
				return
			}
			sites[site.Id] = *param.GetMetadata()
		case model.TYPE_BLOG:
			param, ok := site.RawParams.(*config.BlogConfig)
			if !ok {
				return
			}
			sites[site.Id] = *param.GetMetadata()
		case model.TYPE_SCRAPBOX:
			param, ok := site.RawParams.(*config.ScrapboxConfig)
			if !ok {
				return
			}
			sites[site.Id] = *param.GetMetadata()
		}
	}

	data := map[string]interface{}{
		"Keys":   keys,
		"Posts":  posts,
		"Sites":  sites,
		"Config": hbConfig,
		"Meta": hb.BinMeta{
			Version: GitRevision,
		},
	}
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		slog.Error(fmt.Sprintf("failed to execute template: %s", err))
	}
}
