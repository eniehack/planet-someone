package picker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type MisskeyHandler struct {
	BaseHandler
}

type MisskeyAPIRequestPayload struct {
	UserId       string `json:"userId"`
	WithReplies  bool   `json:"withReplies"`
	WithRenotes  bool   `json:"withRenotes"`
	UntilDate    int    `json:"untilDate"`
	AllowPartial bool   `json:"allowPartial"`
}

type MisskeyAPIResponsePayload struct {
	Id             string  `json:"id"`
	CreatedAt      string  `json:"createdAt"`
	Text           string  `json:"text"`
	ContentWarning *string `json:"cw"`
}

func (h *MisskeyHandler) Pick() error {
	lastRun, err := h.ReadLastRunTime(h.SiteConfig.Id)
	if err != nil {
		log.Println("Error reading last run time:", err)
		lastRun = time.Time{}
	}
	reqUrl, err := url.Parse(h.SiteConfig.SiteUrl)
	if err != nil {
		return fmt.Errorf("cannot parse url:", err)
	}
	reqUrl.Path = "/api/users/notes"
	resp, err := h.Fetch(reqUrl, lastRun)
	if err != nil {
		return fmt.Errorf("cannot fetch misskey posts:", err)
	}
	stmt, err := h.DB.Prepare("INSERT INTO posts (id, title, url, src, date) VALUES (?, ?, ?, ?, ?);")
	if err != nil {
		return fmt.Errorf("cannot make prepare statement: %s", err)
	}
	// 新しい記事を探す
	for _, item := range *resp {
		published, err := time.Parse(time.RFC3339, item.CreatedAt)
		if err != nil {
			log.Println("", err)
			continue
		}
		if 0 <= published.Compare(lastRun) && item.ContentWarning == nil {
			id := BuildID(&published)
			link := fmt.Sprintf("https://%s/notes/%s", reqUrl.Host, item.Id)
			if _, err := stmt.Exec(id, item.Text, link, h.SiteConfig.Id, published.Format(time.RFC3339)); err != nil {
				return fmt.Errorf("cannot insert item(%s): %s", link, err)
			}
		}
	}
	// 現在の時刻を保存
	if err = h.SaveLastRunTime(time.Now(), h.SiteConfig.Id); err != nil {
		return fmt.Errorf("error saving last run time: %s", err)
	}
	return nil
}

func (h *MisskeyHandler) Fetch(reqUrl *url.URL, lastRun time.Time) (*[]MisskeyAPIResponsePayload, error) {
	reqPayload := &MisskeyAPIRequestPayload{
		UserId:       h.SiteConfig.SourceUrl,
		WithReplies:  false,
		WithRenotes:  false,
		UntilDate:    int(lastRun.Unix()),
		AllowPartial: true,
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(reqPayload); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, reqUrl.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("error access Misskey API: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respPayload := []MisskeyAPIResponsePayload{}
	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return nil, err
	}
	return &respPayload, nil
}
