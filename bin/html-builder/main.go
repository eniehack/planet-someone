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
	Id        string
	Content   string
	Url       string
	Date      time.Time
	SiteUrl   string
	SiteTitle string
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
			`SELECT P.id, P.title, P.url, P.date, S.site_url, S.name
			 FROM posts AS P
			 JOIN sources AS S
			   ON P.posts_source = S.id
			 WHERE DATE(date) = ?;`,
			dateStr,
		)
		if err != nil {
			log.Fatalln(err)
		}
		for res.Next() {
			post := Post{}
			if err := res.Scan(&post.Id, &post.Content, &post.Url, &post.Date, &post.SiteUrl, &post.SiteTitle); err != nil {
				log.Fatalln(err)
			}
			posts[dateStr] = append(posts[dateStr], post)
		}
	}
	keys := make([]string, 0, len(posts))
	for k := range posts {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	data := map[string]interface{}{
		"Keys":  keys,
		"Posts": posts,
	}
	if err := tmpl.Execute(os.Stdout, data); err != nil {
		log.Fatalln(err)
	}
}
