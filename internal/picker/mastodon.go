package picker

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
)

type MastodonUserStatusAPIResponse struct {
	Id        string `json:"id"`
	Sensitive bool   `json:"sensitive"`
	CreatedAt string `json:"created_at"`
	Url       string `json:"url"`
	Content   string `json:"content"`
}

type MastodonHandler struct {
	BaseHandler
}

func (h *MastodonHandler) Pick() error {
	lastRun, err := h.ReadLastRunTime(&DEFAULT_DURATION)
	if err != nil {
		slog.Info(fmt.Sprintf("Error reading last run time: %s", err))
	}
	resp, err := h.Fetch()
	if err != nil {
		return err
	}
	stmt, err := h.DB.Prepare("INSERT INTO posts (id, title, url, src, date) VALUES (?, ?, ?, ?, ?);")
	if err != nil {
		return fmt.Errorf("cannot make prepare statement: %s", err)
	}
	// 新しい記事を探す
	for _, item := range *resp {
		published, err := time.Parse(time.RFC3339, item.CreatedAt)
		if err != nil {
			log.Println("mastodon, cannot parse time:", err)
			continue
		}
		fmt.Printf("lastrun: %s, published: %s\n", lastRun.Format(time.RFC3339), item.CreatedAt)
		if lastRun.UnixMilli() < published.UnixMilli() && !item.Sensitive {
			id := BuildID(&published)
			content := buildContent(item.Content)
			if _, err := stmt.Exec(id, content, item.Url, h.SiteConfig.Id, published.UnixMicro()); err != nil {
				return fmt.Errorf("cannot insert item(%s): %s", item.Url, err)
			}
		}
	}
	return nil
}

func buildContent(rawContent string) string {
	contentDoc, err := htmlquery.Parse(strings.NewReader(rawContent))
	if err != nil {
		log.Println("cannot parse html:", err)
	}
	content := new(strings.Builder)
	for _, elem := range htmlquery.Find(contentDoc, "//text()") {
		content.WriteString(htmlquery.InnerText(elem))
	}
	return content.String()
}

func (h MastodonHandler) Fetch() (*[]MastodonUserStatusAPIResponse, error) {
	reqUrl, err := url.Parse(h.SiteConfig.SourceUrl)
	if err != nil {
		return nil, err
	}
	query := make(url.Values)
	query.Add("exclude_replies", "true")
	query.Add("exclude_reblogs", "true")
	reqUrl.RawQuery = query.Encode()
	client := new(http.Client)
	req, err := http.NewRequest(http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error access Misskey API: %s", err)
	}
	defer resp.Body.Close()

	respPayload := []MastodonUserStatusAPIResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return nil, err
	}
	return &respPayload, nil
}
