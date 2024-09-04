package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"sort"
	"time"

	_ "time/tzdata"

	"github.com/eniehack/planet-someone/internal/config"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Post struct {
	Id         string
	Content    string
	Url        string
	Date       string
	ParsedDate *time.Time
	Src        string
}

type Site struct {
	Id      int    `db:"id"`
	Url     string `db:"site_url"`
	IconUrl string `db:"icon_url"`
	Title   string `db:"name"`
}

type Meta struct {
	Url         string
	Description string
	Title       string
}

type Config struct {
	Meta Meta
}

func main() {
	var configFilePath string
	flag.StringVar(&configFilePath, "config", "./config.yml", "config file")
	c := config.ReadConfig(configFilePath)
	db, err := sqlx.Connect("sqlite", fmt.Sprintf("file:%s", c.DB.DB))
	if err != nil {
		log.Fatalln("cannot open db:", err)
	}
	defer db.Close()
	tmpl, err := template.ParseFiles(fmt.Sprintf("%s/index.html", "./template/"))
	if err != nil {
		log.Fatalln("cannot parse template:", err)
	}
	hbConfig := new(Config)
	hbConfig.Meta = Meta{
		Url:         c.Hb.Url,
		Title:       c.Hb.Meta.Title,
		Description: c.Hb.Meta.Description,
	}
	posts := make(map[string][]Post)
	tz, err := time.LoadLocation(c.Hb.TimeZone)
	if err != nil {
		log.Fatalln("cannot parse timezone:", err)
	}
	today := time.Now().In(tz)
	for i := today; today.Sub(i).Abs().Hours() <= (time.Hour * 24 * 14).Hours(); i = i.Add(time.Hour * -24) {
		dateStr := i.Format("2006-01-02")
		res, err := db.Query(
			`SELECT P.id, P.title, P.url, P.date, P.src
			 FROM posts AS P
			 WHERE DATE(P.date, "localtime") = ?
			 ORDER BY P.date DESC;`,
			dateStr,
		)
		if err != nil {
			log.Fatalln(err)
		}
		for res.Next() {
			post := Post{}
			if err := res.Scan(
				&post.Id,
				&post.Content,
				&post.Url,
				&post.Date,
				&post.Src,
			); err != nil {
				log.Fatalln(err)
			}
			parsedDate, err := time.ParseInLocation(time.RFC3339, post.Date, tz)
			if err != nil {
				log.Fatalln(err)
			}
			post.ParsedDate = &parsedDate
			post.Site = &site
			posts[dateStr] = append(posts[dateStr], post)
		}
	}
	keys := make([]string, 0, len(posts))
	for k := range posts {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	sites := map[string]Site{}
	for _, site := range c.Picker.Sites {
		sites[site.Id] = Site{
			Url:     site.SiteUrl,
			IconUrl: site.IconUrl,
			Title:   site.Name,
		}
	}

	data := map[string]interface{}{
		"Keys":   keys,
		"Posts":  posts,
		"Sites":  sites,
		"Config": hbConfig,
	}
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		log.Fatalln(err)
	}
}
