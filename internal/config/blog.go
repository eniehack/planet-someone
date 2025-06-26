package config

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/antchfx/htmlquery"
	pkgUrl "github.com/eniehack/planet-someone/pkg/url"
)

type BlogConfig struct {
	Id        string
	Type      string
	SourceUrl string `yaml:"source_url"`
	SiteUrl   string `yaml:"site_url"`
	Name      string `yaml:"name"`
	IconUrl   string `yaml:"icon_url"`
}

func (c *BlogConfig) SetId(id string) {
	c.Id = id
}

func (c *BlogConfig) SetType(typ string) {
	c.Type = typ
}

func (c *BlogConfig) GetType() string {
	return c.Type
}

func (c *BlogConfig) CompleteMetadata(ctx context.Context, id string) error {
	if len(c.SiteUrl) == 0 {
		return fmt.Errorf("%s: site_url is undefined", id)
	}
	client := new(http.Client)
	reqUrl, err := url.Parse(c.SiteUrl)
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
	if len(c.Name) == 0 {
		titleElem := htmlquery.FindOne(doc, `//title/text()`)
		c.Name = htmlquery.InnerText(titleElem)
	}
	if len(c.SourceUrl) == 0 {
		feedUrlElem := htmlquery.FindOne(doc, `//link[@rel="alternate" and (@type="application/rss+xml" or @type="application/atom+xml")]/@href`)
		srcUrl, err := pkgUrl.ResolveAbsUrl(reqUrl, htmlquery.InnerText(feedUrlElem))
		if err != nil {
			return err
		}
		c.SourceUrl = srcUrl.String()
	}
	if len(c.IconUrl) == 0 {
		iconUrlElem := htmlquery.FindOne(doc, `//link[@rel="icon"]/@href`)
		iconUrl, err := pkgUrl.ResolveAbsUrl(reqUrl, htmlquery.InnerText(iconUrlElem))
		if err != nil {
			return err
		}
		c.IconUrl = iconUrl.String()
	}
	return nil
}
