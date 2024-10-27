package main

import (
	"flag"
	"fmt"
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
	for _, src := range c.Picker.Sites {
		p, err := picker.PickerFactory(db, &src)
		if err != nil {
			slog.Error(fmt.Sprintf("cannot create picker on pickerfactory: %s", err))
			os.Exit(1)
		}
		if err := p.Pick(); err != nil {
			slog.Warn(fmt.Sprintf("cannot pick posts: %s", err))
		}
		time.Sleep(time.Second * 1)
	}
}
