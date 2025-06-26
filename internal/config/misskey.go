package config

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/antchfx/htmlquery"
	pkgUrl "github.com/eniehack/planet-someone/pkg/url"
)

type MisskeyConfig struct {
	Id          string
	Type        string
	InstanceUrl string `yaml:"instance_url"`
	UserId      string `yaml:"user_id"`
	SiteUrl     string `yaml:"site_url"`
	Name        string `yaml:"name"`
	IconUrl     string `yaml:"icon_url"`
}

func (c *MisskeyConfig) SetId(id string) {
	c.Id = id
}

func (c *MisskeyConfig) SetType(typ string) {
	c.Type = typ
}

func (c *MisskeyConfig) GetType() string {
	return c.Type
}

func (c *MisskeyConfig) CompleteMetadata(ctx context.Context, id string) error {
	if len(c.InstanceUrl) == 0 {
		return errors.New("instance_url is required")
	}
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
