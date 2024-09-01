package picker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/oklog/ulid/v2"
)

type MisskeyHandler struct {
	BaseHandler
}

type APIRequestPayload struct {
	UserId       string `json:"userId"`
	WithReplies  bool   `json:"withReplies"`
	WithRenotes  bool   `json:"withRenotes"`
	UntilDate    int    `json:"untilDate"`
	AllowPartial bool   `json:"allowPartial"`
}

type APIResponsePayload struct {
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
	reqUrl.Path = "/api/users/notes"
	if err != nil {
		return err
	}
	reqPayload := &APIRequestPayload{
		UserId:       h.SiteConfig.SourceUrl,
		WithReplies:  false,
		WithRenotes:  false,
		UntilDate:    int(lastRun.UnixMilli()),
		AllowPartial: true,
	}
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(reqPayload); err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, reqUrl.String(), buf)
	if err != nil {
		return fmt.Errorf("error access Misskey API: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respPayload := []APIResponsePayload{}
	if err := json.NewDecoder(resp.Body).Decode(&respPayload); err != nil {
		return err
	}
	stmt, err := h.DB.Prepare("INSERT INTO posts (id, title, url, src, date) VALUES (?, ?, ?, ?, ?);")
	if err != nil {
		return fmt.Errorf("cannot make prepare statement: %s", err)
	}
	// 新しい記事を探す
	for _, item := range respPayload {
		published, err := time.Parse(time.RFC3339, item.CreatedAt)
		if err != nil {
			log.Println("", err)
			continue
		}
		if 0 <= published.Compare(lastRun) && item.ContentWarning == nil {
			id := ulid.Make()
			id.SetTime(uint64(published.UnixMilli()))
			link := fmt.Sprintf("https://%s/notes/%s", reqUrl.Host, item.Id)
			if _, err := stmt.Exec(id.String(), item.Text, link, h.SiteConfig.Id, published.Format(time.RFC3339)); err != nil {
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
