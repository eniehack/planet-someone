package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/eniehack/planet-eniehack/internal/config"
	"github.com/eniehack/planet-eniehack/internal/model"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	_ "modernc.org/sqlite"
)

const SQLITE = "sqlite"

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
	db, err := sqlx.Connect(SQLITE, fmt.Sprintf("file:%s", c.DB.DB))
	if err != nil {
		log.Fatalln("cannot connect to sqlite file: ", err)
	}
	defer db.Close()
	ctx := context.Background()
	migrations := &migrate.FileMigrationSource{
		Dir: c.DB.MigrationDir,
	}
	n, err := migrate.ExecContext(ctx, db.DB, SQLITE+"3", migrations, migrate.Up)
	if err != nil {
		log.Fatalln("cannot exec migration: ", err)
	}
	log.Printf("apply %d migrations", n)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalln("cannot open transaction: ", err)
	}
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO sources (id_alias, site_url, source_url, name, type) VALUES (?, ?, ?, ?, ?);")
	if err != nil {
		log.Fatalln("cannot create prepare statement:", err)
	}
	for _, site := range c.Picker.Sites {
		if _, err := stmt.ExecContext(ctx, site.Id, site.SiteUrl, site.SourceUrl, site.Name, model.LookupTypeNumber(site.Type)); err != nil {
			tx.Rollback()
			log.Fatalln("cannot exec statement on inserting site config:", err)
		}
	}
	if err := tx.Commit(); err != nil {
		log.Fatalln("cannot commit tx:", err)
	}
}
