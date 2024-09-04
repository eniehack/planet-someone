package picker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
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
	UntilDate    int64  `json:"untilDate"`
	AllowPartial bool   `json:"allowPartial"`
}

type MisskeyAPIResponsePayload struct {
	Id             string  `json:"id"`
	CreatedAt      string  `json:"createdAt"`
	Text           string  `json:"text"`
	ContentWarning *string `json:"cw"`
}

func (h *MisskeyHandler) Pick() error {
	lastRun, err := h.ReadLastRunTime(&DEFAULT_DURATION)
	if err != nil {
		slog.Info(fmt.Sprintf("Error reading last run time: %s", err))
	}
	reqUrl, err := url.Parse(h.SiteConfig.SiteUrl)
	if err != nil {
		return fmt.Errorf("cannot parse url: %s", err)
	}
	reqUrl.Path = "/api/users/notes"
	resp, err := h.Fetch(reqUrl, lastRun)
	if err != nil {
		return fmt.Errorf("cannot fetch misskey posts: %s", err)
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
		fmt.Printf("lastrun: %s, published: %s\n", lastRun.Format(time.RFC3339), item.CreatedAt)
		if lastRun.UnixMilli() < published.UnixMilli() && item.ContentWarning == nil {
			id := BuildID(&published)
			link := fmt.Sprintf("https://%s/notes/%s", reqUrl.Host, item.Id)
			if _, err := stmt.Exec(id, item.Text, link, h.SiteConfig.Id, published.UnixMicro()); err != nil {
				return fmt.Errorf("cannot insert item(%s): %s", link, err)
			}
		}
	}
	return nil
}

func (h *MisskeyHandler) Fetch(reqUrl *url.URL, lastRun *time.Time) (*[]MisskeyAPIResponsePayload, error) {
	reqPayload := &MisskeyAPIRequestPayload{
		UserId:       h.SiteConfig.SourceUrl,
		WithReplies:  false,
		WithRenotes:  false,
		UntilDate:    lastRun.UnixMilli(),
		AllowPartial: true,
	}
	slog.Debug("req params to misskey", "val", reqPayload)
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(reqPayload); err != nil {
		slog.Error("misskey: json encode error")
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
		slog.Error("misskey: http req error")
		return nil, err
	}
	defer resp.Body.Close()
	slog.Debug("misskey resp status:", "val", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected statuscode")
	}
	respBuf := new(bytes.Buffer)
	respBuf.ReadFrom(resp.Body)
	slog.Debug("resp from misskey", "val", respBuf.String())

	respPayload := []MisskeyAPIResponsePayload{}
	if err := json.NewDecoder(respBuf).Decode(&respPayload); err != nil {
		slog.Error("misskey: json decode error")
		return nil, err
	}
	return &respPayload, nil
}
