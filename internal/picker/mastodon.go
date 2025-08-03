package picker

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/eniehack/planet-someone/internal/config"
	"golang.org/x/net/html"
)

type MastodonAccountAPIResponse struct {
	Indexable bool   `json:"indexable"`
	Acct      string `json:"acct"`
}

type MastodonUserStatusAPIResponse struct {
	Id        string                         `json:"id"`
	Sensitive bool                           `json:"sensitive"`
	CreatedAt string                         `json:"created_at"`
	Url       string                         `json:"url"`
	Content   string                         `json:"content"`
	Account   *MastodonAccountAPIResponse    `json:"account"`
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
	stmt, err := h.DB.Prepare("INSERT INTO posts (id, content, url, src, type, created_at) VALUES (?, ?, ?, ?, ?, ?);")
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
			var content string
			if item.Reblog == nil {
				node, err := html.Parse(strings.NewReader(item.Content))
				if err != nil {
					return err
				}
				content = buildContent(node, true)
				if _, err := stmt.Exec(id, content, item.Url, h.Config.Id, h.Config.Type, published.Unix()); err != nil {
					return fmt.Errorf("cannot insert item(%s): %s", item.Url, err)
				}
			} else {
				node, err := html.Parse(strings.NewReader(item.Reblog.Content))
				if err != nil {
					return err
				}
				content = fmt.Sprintf("BT %s: %s", item.Reblog.Account.Acct, buildContent(node, true))
				if _, err := stmt.Exec(id, content, item.Reblog.Url, h.Config.Id, h.Config.Type, published.Unix()); err != nil {
					return fmt.Errorf("cannot insert item(%s): %s", item.Url, err)
				}
			}
		}
	}
	return nil
}

func buildContent(node *html.Node, insertContentFlag bool) string {
	var content strings.Builder

	if node.Type == html.ElementNode {
		switch node.Data {
		case "a":
			cls := getClassList(node)
			if !slices.Contains(cls, "hashtag") {
				if href := getAttribute(node, "href"); href != "" {
					content.WriteString(href)
					insertContentFlag = false
				}
			} else {
				insertContentFlag = true
			}
		case "br":
			content.WriteString("\n")
		case "span":
			cls := getClassList(node)
			if slices.Contains(cls, "invisible") || slices.Contains(cls, "ellipsis") {
				insertContentFlag = false
			}
		}
	}

	if node.Type == html.TextNode && insertContentFlag {
		content.WriteString(node.Data + " ")
	}

	// 子ノードを再帰的に処理
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		content.WriteString(buildContent(child, insertContentFlag))
	}

	return content.String()
}

// ヘルパー関数: class属性を取得してスライスに変換
func getClassList(node *html.Node) []string {
	for _, attr := range node.Attr {
		if attr.Key == "class" {
			return strings.Fields(attr.Val)
		}
	}
	return []string{}
}

// ヘルパー関数: 指定した属性値を取得
func getAttribute(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func (h MastodonHandler) Fetch() (*[]MastodonUserStatusAPIResponse, error) {
	reqUrl, err := url.Parse(h.Config.SourceUrl)
	if err != nil {
		return nil, err
	}
	fmt.Println(reqUrl.String())
	query := make(url.Values)
	query.Add("exclude_replies", "true")
	//query.Add("exclude_reblogs", "true")
	reqUrl.RawQuery = query.Encode()
	fmt.Println(reqUrl.String())
	client := new(http.Client)
	req, err := http.NewRequest(http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", config.UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error access Mastodon API: %s", err)
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
