package picker

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/eniehack/planet-someone/internal/config"
)

type MisskeyHandler struct {
	BaseHandler
	Config *config.MisskeyConfig
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
	reqUrl, err := url.Parse(h.Config.InstanceUrl)
	if err != nil {
		return fmt.Errorf("cannot parse url: %s", err)
	}
	reqUrl.Path = "/api/users/notes"
	resp, err := h.Fetch(reqUrl, lastRun)
	if err != nil {
		return fmt.Errorf("cannot fetch misskey posts: %s", err)
	}
	stmt, err := h.DB.Prepare("INSERT INTO posts (id, content, url, src, type, created_at) VALUES (?, ?, ?, ?, ?, ?);")
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
		if lastRun.Unix() < published.Unix() && item.ContentWarning == nil {
			id := BuildID(&published)
			link := fmt.Sprintf("https://%s/notes/%s", reqUrl.Host, item.Id)
			if _, err := stmt.Exec(id, item.Text, link, h.Config.Id, h.Config.Type, published.Unix()); err != nil {
				return fmt.Errorf("cannot insert item(%s): %s", link, err)
			}
		}
	}
	return nil
}

func (h *MisskeyHandler) Fetch(reqUrl *url.URL, lastRun *time.Time) (*[]MisskeyAPIResponsePayload, error) {
	reqPayload := &MisskeyAPIRequestPayload{
		UserId:       h.Config.UserId,
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
	req.Header.Set("User-Agent", config.UserAgent)
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

func (h *MisskeyHandler) ReadLastRunTime(dur *time.Duration) (*time.Time, error) {
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
