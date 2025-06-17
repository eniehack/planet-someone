package picker

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/eniehack/planet-someone/internal/config"
)

type MastodonUserStatusAPIResponse struct {
	Id        string                         `json:"id"`
	Sensitive bool                           `json:"sensitive"`
	CreatedAt string                         `json:"created_at"`
	Url       string                         `json:"url"`
	Content   string                         `json:"content"`
	Reblog    *MastodonUserStatusAPIResponse `json:"reblog,omitempty"`
}

type MastodonHandler struct {
	BaseHandler
	Config *config.MastodonConfig
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
	stmt, err := h.DB.Prepare("INSERT INTO posts (id, title, url, src, created_at) VALUES (?, ?, ?, ?, ?);")
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
		if lastRun.Unix() < published.Unix() && !item.Sensitive {
			id := BuildID(&published)
			content := buildContent(item.Content)
			if _, err := stmt.Exec(id, content, item.Url, h.Config.Id, published.Unix()); err != nil {
				return fmt.Errorf("cannot insert item(%s): %s", item.Url, err)
			}
		}
	}
	return nil
}

func buildContent(rawContent string) string {
	brReplacedContent := strings.ReplaceAll(rawContent, "<br />", "\n")
	contentDoc, err := htmlquery.Parse(strings.NewReader(brReplacedContent))
	if err != nil {
		log.Println("cannot parse html:", err)
	}
	content := new(strings.Builder)
	for _, elem := range htmlquery.Find(contentDoc, "//text()") {
		content.WriteString(htmlquery.InnerText(elem) + " ")
	}
	return content.String()
}

func (h MastodonHandler) Fetch() (*[]MastodonUserStatusAPIResponse, error) {
	reqUrl, err := url.Parse(h.Config.SourceUrl)
	if err != nil {
		return nil, err
	}
	query := make(url.Values)
	query.Add("exclude_replies", "true")
	//query.Add("exclude_reblogs", "true")
	reqUrl.RawQuery = query.Encode()
	client := new(http.Client)
	req, err := http.NewRequest(http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", config.UserAgent)
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

func (h *MastodonHandler) ReadLastRunTime(dur *time.Duration) (*time.Time, error) {
	row := h.DB.QueryRow("SELECT created_at FROM posts WHERE src = ? ORDER BY created_at DESC;", h.Config.Id)
	if row.Err() != nil {
		if row.Err() == sql.ErrNoRows {
			t := time.Now().Add(*dur)
			return &t, row.Err()
		}
		return nil, row.Err()
	}
	var timestamp_unit int64
	row.Scan(&timestamp_unit)
	timestamp := time.Unix(timestamp_unit, 0)
	return &timestamp, nil
}
