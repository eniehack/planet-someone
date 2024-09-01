package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"sort"
	"time"

	"github.com/eniehack/planet-eniehack/internal/config"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Post struct {
	Id          string
	Content     string
	Url         string
	Date        string
	ParsedDate  *time.Time
	SiteId      string
	SiteUrl     string
	SiteIconUrl string
	SiteTitle   string
}

type Site struct {
	Id      int    `db:"id"`
	Url     string `db:"site_url"`
	IconUrl string `db:"icon_url"`
	Title   string `db:"name"`
}

func readConfig(configFilePath string) *config.Config {
	f, err := os.Open(configFilePath)
	if err != nil {
		log.Fatalln("cannot open config file:", err)
	}
	defer f.Close()
	c, err := config.New(f)
	if err != nil {
		log.Fatalln("cannot parse config file:", err)
	}
	return c
}

func main() {
	var configFilePath string
	flag.StringVar(&configFilePath, "config", "./config.yml", "config file")
	c := readConfig(configFilePath)
	db, err := sqlx.Connect("sqlite", fmt.Sprintf("file:%s", c.DB.DB))
	if err != nil {
		log.Fatalln("cannot open db:", err)
	}
	defer db.Close()
	tmpl, err := template.ParseFiles(fmt.Sprintf("%s/index.html", "./template/"))
	if err != nil {
		log.Fatalln("cannot parse template:", err)
	}
	posts := make(map[string][]Post)
	today := time.Now()
	for i := today; today.Sub(i).Abs().Hours() <= (time.Hour * 24 * 14).Hours(); i = i.Add(time.Hour * -24) {
		dateStr := i.Format("2006-01-02")
		res, err := db.Query(
			`SELECT P.id, P.title, P.url, P.date, S.id, S.site_url, S.icon_url, S.name
			 FROM posts AS P
			 JOIN sources AS S
			   ON P.src = S.id
			 WHERE DATE(date) = ?;`,
			dateStr,
		)
		if err != nil {
			log.Fatalln(err)
		}
		for res.Next() {
			post := Post{}
			if err := res.Scan(&post.Id, &post.Content, &post.Url, &post.Date, &post.SiteId, &post.SiteUrl, &post.SiteIconUrl, &post.SiteTitle); err != nil {
				log.Fatalln(err)
			}
			parsedDate, err := time.Parse(time.RFC3339, post.Date)
			if err != nil {
				log.Println("cannot parse date:", err)
			}
			post.ParsedDate = &parsedDate
			posts[dateStr] = append(posts[dateStr], post)
		}
	}
	keys := make([]string, 0, len(posts))
	for k := range posts {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	sites := []Site{}

	if err := db.Select(&sites, "SELECT id, site_url, icon_url, name from sources;"); err != nil {
		log.Fatalln(err)
	}
	data := map[string]interface{}{
		"Keys":  keys,
		"Posts": posts,
		"Sites": sites,
	}
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		log.Fatalln(err)
	}
}
