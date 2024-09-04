package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/eniehack/planet-someone/internal/config"
	"github.com/eniehack/planet-someone/internal/picker"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func main() {
	var configFilePath string
	flag.StringVar(&configFilePath, "config", "./config.yml", "config file")
	c := config.ReadConfig(configFilePath)
	db, err := sqlx.Connect("sqlite", fmt.Sprintf("file:%s", c.DB.DB))
	if err != nil {
		log.Fatalln("cannot open db:", err)
	}
	defer db.Close()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	for _, src := range c.Picker.Sites {
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
