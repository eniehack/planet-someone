package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/eniehack/planet-someone/internal/config"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
	_ "modernc.org/sqlite"
)

const (
	SQLITE = "sqlite"
)

func resolveAbsUrl(baseUrl *url.URL, path string) (*url.URL, error) {
	relUrl, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	abs := baseUrl.ResolveReference(relUrl)
	return abs, nil
}

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
						Action: func(ctx context.Context, cmd *cli.Command) error {
							c := config.ReadConfig(cmd.String("config"))
							newSites := []config.SiteConfig{}
							for _, siteConfig := range c.Picker.Sites {
								if len(siteConfig.Id) == 0 {
									return errors.New("id is required")
								}
								if len(siteConfig.SiteUrl) == 0 {
									return fmt.Errorf("%s: site_url is undefined", siteConfig.Id)
								}
								client := new(http.Client)
								reqUrl, err := url.Parse(siteConfig.SiteUrl)
								if err != nil {
									return err
								}
								req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl.String(), nil)
								if err != nil {
									return err
								}
								req.Header.Set("User-Agent", config.UserAgent)
								resp, err := client.Do(req)
								if err != nil {
									return err
								}
								defer resp.Body.Close()
								doc, err := htmlquery.Parse(resp.Body)
								if err != nil {
									return err
								}
								if len(siteConfig.Name) == 0 {
									titleElem := htmlquery.FindOne(doc, `//title/text()`)
									siteConfig.Name = htmlquery.InnerText(titleElem)
								}
								if len(siteConfig.SourceUrl) == 0 {
									feedUrlElem := htmlquery.FindOne(doc, `//link[@rel="alternate" and (@type="application/rss+xml" or @type="application/atom+xml")]/@href`)
									srcUrl, err := resolveAbsUrl(reqUrl, htmlquery.InnerText(feedUrlElem))
									if err != nil {
										return err
									}
									siteConfig.SourceUrl = srcUrl.String()
								}
								if len(siteConfig.IconUrl) == 0 {
									iconUrlElem := htmlquery.FindOne(doc, `//link[@rel="icon"]/@href`)
									iconUrl, err := resolveAbsUrl(reqUrl, htmlquery.InnerText(iconUrlElem))
									if err != nil {
										return err
									}
									siteConfig.IconUrl = iconUrl.String()
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
						},
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
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							if cmd.Args().Len() != 1 {
								return errors.New("must be 1 argument")
							}
							c := config.ReadConfig(cmd.String("config"))
							c.Picker.Sites = append(c.Picker.Sites, config.SiteConfig{
								Id: cmd.Args().First(),
							})
							f, err := os.OpenFile(cmd.String("config"), os.O_WRONLY|os.O_TRUNC, 0644)
							if err != nil {
								return err
							}
							defer f.Close()
							if err := yaml.NewEncoder(f).Encode(c); err != nil {
								return err
							}
							return nil
						},
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
						Action: func(ctx context.Context, cmd *cli.Command) error {
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
						},
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
