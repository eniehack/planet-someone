package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/antchfx/htmlquery"
	"github.com/eniehack/planet-eniehack/internal/model"
	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/urfave/cli/v3"
	_ "modernc.org/sqlite"
)

const (
	SQLITE    = "sqlite"
	UserAgent = "Mozilla/5.0 (compatible; planet-eniehack; +https://github.com/eniehack/planet-someone)"
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
	db, err := sqlx.Connect(SQLITE, fmt.Sprintf("file:%s", cmd.String("database")))
	if err != nil {
		return fmt.Errorf("cannot connect to sqlite file: %s", err)
	}
	defer db.Close()
	migrations := &migrate.FileMigrationSource{
		Dir: cmd.String("migration_dir"),
	}
	n, err := migrate.ExecContext(ctx, db.DB, SQLITE+"3", migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("cannot exec migration: %s", err)
	}
	log.Printf("apply %d migrations", n)
	return nil
}

type SiteMetadata struct {
	Title   string `goquery:"title,text"`
	IconUrl string `goquery:"link,[rel=icon],[href]"`
	FeedUrl string `goquery:"link,[rel=alternate]"`
}

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name: "site",
				Commands: []*cli.Command{
					{
						Name:  "add",
						Usage: "add site",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "type",
								Value: "blog",
							},
							&cli.StringFlag{
								Name:     "site-url",
								Aliases:  []string{"url"},
								Required: true,
							},
							&cli.StringFlag{
								Name:     "source-url",
								Aliases:  []string{"src"},
								Required: false,
							},
							&cli.StringFlag{
								Name:     "database",
								Aliases:  []string{"db"},
								Required: false,
							},
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							db, err := sqlx.Connect(SQLITE, fmt.Sprintf("file:%s", cmd.String("database")))
							if err != nil {
								return fmt.Errorf("cannot connect to sqlite file: %s", err)
							}
							defer db.Close()
							client := new(http.Client)
							reqUrl, err := url.Parse(cmd.String("site-url"))
							if err != nil {
								return err
							}
							req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl.String(), nil)
							if err != nil {
								return err
							}
							req.Header.Set("User-Agent", UserAgent)
							resp, err := client.Do(req)
							if err != nil {
								return err
							}
							defer resp.Body.Close()
							doc, err := htmlquery.Parse(resp.Body)
							if err != nil {
								return err
							}
							iconUrlElem := htmlquery.FindOne(doc, `//link[@rel="icon"]/@href`)
							feedUrlElem := htmlquery.FindOne(doc, `//link[@rel="alternate" and (@type="application/rss+xml" or @type="application/atom+xml")]/@href`)
							titleElem := htmlquery.FindOne(doc, `//title/text()`)
							iconUrl, err := resolveAbsUrl(reqUrl, htmlquery.InnerText(iconUrlElem))
							if err != nil {
								return err
							}
							title := htmlquery.InnerText(titleElem)
							var srcUrl *url.URL
							if len(cmd.String("source-url")) != 0 {
								srcUrl, err = resolveAbsUrl(reqUrl, cmd.String("source-url"))
								if err != nil {
									return nil
								}
							} else if feedUrlElem != nil {
								srcUrl, err = resolveAbsUrl(reqUrl, htmlquery.InnerText(feedUrlElem))
								if err != nil {
									return nil
								}
							} else {
								return fmt.Errorf("srcUrl is empty")
							}
							tx, err := db.BeginTx(ctx, nil)
							if err != nil {
								return fmt.Errorf("cannot open transaction: %s", err)
							}
							if _, err := tx.ExecContext(
								ctx,
								`INSERT INTO sources (site_url, source_url, icon_url, name, type) VALUES (?, ?, ?, ?, ?);`,
								reqUrl.String(),
								srcUrl.String(),
								iconUrl.String(),
								title,
								model.LookupTypeNumber(cmd.String("type")),
							); err != nil {
								tx.Rollback()
								return fmt.Errorf("cannot exec statement on inserting site config: %s", err)
							}
							if err := tx.Commit(); err != nil {
								log.Fatalln("cannot commit tx:", err)
							}
							return nil
						},
					},
					{
						Name:    "list",
						Usage:   "list site",
						Aliases: []string{"ls"},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							return nil
						},
					},
					{
						Name:    "remove",
						Usage:   "rm site",
						Aliases: []string{"rm"},
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name: "id",
							},
						},
						Action: func(ctx context.Context, cmd *cli.Command) error {
							return nil
						},
					},
				},
			},
			{
				Name: "init",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "database",
						Aliases: []string{"db"},
						Value:   "./planet.sqlite",
					},
					&cli.StringFlag{
						Name:    "migration_dir",
						Aliases: []string{"m"},
						Value:   "./db/migrations",
					},
				},
				Action: initAction,
			},
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatalln(err)
	}
}
