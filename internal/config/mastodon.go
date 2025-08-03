package config

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/antchfx/htmlquery"
	"github.com/eniehack/planet-someone/internal/hb"
	pkgUrl "github.com/eniehack/planet-someone/pkg/url"
)

type MastodonConfig struct {
	Id        string
	Type      string
	SourceUrl string `yaml:"source_url"`
	SiteUrl   string `yaml:"site_url"`
	Name      string `yaml:"name"`
	IconUrl   string `yaml:"icon_url"`
	Acct      string `yaml:"acct"`
}

type MastodonAccountLookupAPIResponse struct {
	Id string `yaml:"id"`
}

func (c *MastodonConfig) SetId(id string) {
	c.Id = id
}

func (c *MastodonConfig) SetType(typ string) {
	c.Type = typ
}

func (c *MastodonConfig) GetType() string {
	return c.Type
}

func (c *MastodonConfig) GetMetadata() *hb.Site {
	return &hb.Site{
		Url:     c.SiteUrl,
		IconUrl: c.IconUrl,
		Title:   c.Name,
	}
}

func (c *MastodonConfig) CompleteMetadata(ctx context.Context, id string) error {
	if len(c.SiteUrl) == 0 {
		return fmt.Errorf("%s: site_url is undefined", id)
	}
	if len(c.Acct) == 0 {
		return fmt.Errorf("%s: acct is undefined", id)
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
		lookupEndpoint, _ := url.Parse(c.SiteUrl)
		lookupEndpoint.Path = "/api/v1/accounts/lookup"
		query := lookupEndpoint.Query()
		query.Add("acct", c.Acct)
		lookupEndpoint.RawQuery = query.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, lookupEndpoint.String(), nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", UserAgent)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		lookupRespPayload := new(MastodonAccountLookupAPIResponse)
		if err := json.NewDecoder(resp.Body).Decode(lookupRespPayload); err != nil {
			return err
		}
		srcUrl, _ := url.Parse(c.SiteUrl)
		srcUrl.Path = fmt.Sprintf("/api/v1/accounts/%s/statuses", lookupRespPayload.Id)
		fmt.Println(srcUrl.String())
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
