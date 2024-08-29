package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eniehack/planet-eniehack/internal/config"
	"github.com/eniehack/planet-eniehack/internal/picker"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func readConfig(configFilePath string) *config.Config {
	f, err := os.Open(configFilePath)
	if err != nil {
		log.Fatalf("cannot open config file: %s", err)
	}
	defer f.Close()
	c, err := config.New(f)
	if err != nil {
		log.Fatalf("cannot parse config file: %s", err)
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
	srcs := []picker.Source{}
	if err := db.Select(&srcs, "SELECT id, id_alias, source_url, type FROM sources;"); err != nil {
		log.Fatalln("cannot exec query fetch sources:", err)
	}
	for _, src := range srcs {
		p, err := picker.PickerFactory(db, &src)
		if err != nil {
			log.Fatalln("cannot create picker on pickerfactory:", err)
		}
		if err := p.Pick(); err != nil {
			log.Fatalln(err)
		}
		time.Sleep(time.Second * 1)
	}
}
