package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eniehack/planet-someone/internal/config"
	"github.com/eniehack/planet-someone/internal/model"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
	_ "modernc.org/sqlite"
)

const (
	SQLITE = "sqlite"
)

func initAction(ctx context.Context, cmd *cli.Command) error {
	c := config.ReadConfig(cmd.String("config"))
	db, err := sqlx.Connect(SQLITE, fmt.Sprintf("file:%s", c.DB.DB))
	if err != nil {
		return fmt.Errorf("cannot connect to sqlite file: %s", err)
	}
	defer db.Close()
	migrations := &migrate.FileMigrationSource{
		Dir: c.DB.MigrationDir,
	}
	n, err := migrate.ExecContext(ctx, db.DB, SQLITE+"3", migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("cannot exec migration: %s", err)
	}
	log.Printf("apply %d migrations", n)
	return nil
}

func validateConfig(ctx context.Context, cmd *cli.Command) error {
	c := config.ReadConfig(cmd.String("config"))
	newSites := []config.SiteConfigWrapper{}
	for _, siteConfig := range c.Picker.Sites {
		if len(siteConfig.Id) == 0 {
			return errors.New("id is required")
		}
		params, err := config.GetParam(siteConfig)
		if err != nil {
			return err
		}
		if err := params.CompleteMetadata(ctx, siteConfig.Id); err != nil {
			return err
		}
		time.Sleep(time.Second * 1)
		newSites = append(newSites, siteConfig)
	}
	f, err := os.OpenFile(cmd.String("config"), os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	nc := c
	nc.Picker.Sites = newSites
	if err := yaml.NewEncoder(f).Encode(nc); err != nil {
		return err
	}
	return nil
}

func addSite(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() != 1 {
		return errors.New("must be 1 argument")
	}
	c := config.ReadConfig(cmd.String("config"))
	var scw *config.SiteConfigWrapper
	switch cmd.String("type") {
	case model.TYPE_MASTODON:
		scw = &config.SiteConfigWrapper{
			Type:      cmd.String("type"),
			Id:        cmd.Args().First(),
			RawParams: &config.MastodonConfig{},
		}
	case model.TYPE_MISSKEY:
		scw = &config.SiteConfigWrapper{
			Type:      cmd.String("type"),
			Id:        cmd.Args().First(),
			RawParams: &config.MisskeyConfig{},
		}
	case model.TYPE_SCRAPBOX:
		scw = &config.SiteConfigWrapper{
			Type:      cmd.String("type"),
			Id:        cmd.Args().First(),
			RawParams: &config.ScrapboxConfig{},
		}
	case model.TYPE_BLOG:
		scw = &config.SiteConfigWrapper{
			Type:      cmd.String("type"),
			Id:        cmd.Args().First(),
			RawParams: &config.BlogConfig{},
		}
	default:
		return errors.New("unexpected site type")
	}
	c.Picker.Sites = append(c.Picker.Sites, *scw)

	f, err := os.OpenFile(cmd.String("config"), os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := yaml.NewEncoder(f).Encode(c); err != nil {
		return err
	}
	return nil
}

func removeSite(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() != 1 {
		return errors.New("must be 1 argument")
	}
	c := config.ReadConfig(cmd.String("config"))
	db, err := sqlx.Connect(SQLITE, fmt.Sprintf("file:%s", c.DB.DB))
	if err != nil {
		return fmt.Errorf("cannot connect to sqlite file: %s", err)
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, "DELETE FROM posts WHERE src = ?;", cmd.Args().First()); err != nil {
		return fmt.Errorf("cannot delete posts: %s", err)
	}
	return nil
}

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name: "site",
				Commands: []*cli.Command{
					{
						Name:  "validate",
						Usage: "validate config file",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "config",
								Aliases: []string{"c"},
							},
						},
						Action: validateConfig,
					},
					{
						Name:      "add",
						ArgsUsage: "[id]",
						Usage:     "add source data to db",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "config",
								Aliases: []string{"c"},
							},
							&cli.StringFlag{
								Name:     "type",
								Aliases:  []string{"t"},
								Required: true,
							},
						},
						Action: addSite,
					},
					{
						Name:      "remove",
						ArgsUsage: "[id]",
						Usage:     "remove posts taken from specific source from db.",
						Aliases:   []string{"rm"},
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "config",
								Aliases: []string{"c"},
							},
						},
						Action: removeSite,
					},
				},
			},
			{
				Name:   "init",
				Action: initAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
					},
				},
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatalln(err)
	}
}
