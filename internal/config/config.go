package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"

	"github.com/eniehack/planet-someone/internal/hb"
	"github.com/eniehack/planet-someone/internal/model"
	"gopkg.in/yaml.v3"
)

const (
	UserAgent = "Mozilla/5.0 (compatible; planet-eniehack; +https://github.com/eniehack/planet-someone)"
)

func New(f io.Reader) (*Config, error) {
	config := new(Config)
	if err := yaml.NewDecoder(f).Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}

func ReadConfig(configFilePath string) *Config {
	f, err := os.Open(configFilePath)
	if err != nil {
		log.Fatalln("cannot open config file:", err)
	}
	defer f.Close()
	c, err := New(f)
	if err != nil {
		log.Fatalln("cannot parse config file:", err)
	}
	return c
}

func (sw *SiteConfigWrapper) UnmarshalYAML(value *yaml.Node) error {
	type rawWrapper struct {
		Type   string      `yaml:"type"`
		Id     string      `yaml:"id"`
		Params interface{} `yaml:"params"` // paramsはtypeによって形が異なるので、一旦interface{}としてDecodeし、paramsだけ再度encodeすることで狙ったstructに変換できる
	}
	raw := new(rawWrapper)
	if err := value.Decode(&raw); err != nil {
		return err
	}
	sw.Type = raw.Type
	sw.Id = raw.Id

	buf := new(bytes.Buffer)
	if err := yaml.NewEncoder(buf).Encode(raw.Params); err != nil {
		return err
	}

	switch raw.Type {
	case model.TYPE_MASTODON:
		cfg := new(MastodonConfig)
		if err := yaml.NewDecoder(buf).Decode(cfg); err != nil {
			return fmt.Errorf("failed to decode mastodon params: %w", err)
		}
		sw.RawParams = cfg
	case model.TYPE_MISSKEY:
		cfg := new(MisskeyConfig)
		if err := yaml.NewDecoder(buf).Decode(cfg); err != nil {
			return fmt.Errorf("failed to decode misskey params: %w", err)
		}
		sw.RawParams = cfg
	case model.TYPE_SCRAPBOX:
		cfg := new(ScrapboxConfig)
		if err := yaml.NewDecoder(buf).Decode(cfg); err != nil {
			return fmt.Errorf("failed to decode scrapbox params: %w", err)
		}
		sw.RawParams = cfg
	case model.TYPE_BLOG:
		cfg := new(BlogConfig)
		if err := yaml.NewDecoder(buf).Decode(cfg); err != nil {
			return fmt.Errorf("failed to decode blog params: %w", err)
		}
		sw.RawParams = cfg
	default:
		return fmt.Errorf("unknown type: %s", raw.Type)
	}

	return nil
}

func (sw *SiteConfigWrapper) MarshalYAML() (interface{}, error) {
	if sw.RawParams == nil {
		return nil, fmt.Errorf("SiteConfig is nil")
	}
	result := map[string]interface{}{
		"type": sw.Type,
		"id":   sw.Id,
	}

	params := map[string]interface{}{}
	v := reflect.ValueOf(sw.RawParams)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("yaml")
		if len(tag) == 0 || tag == "-" {
			continue
		}
		params[tag] = v.Field(i).Interface()
	}
	result["params"] = params
	return result, nil
}

func GetParams[T any](s SiteConfigWrapper) (*T, bool) {
	if params, ok := s.RawParams.(T); ok {
		return &params, true
	}
	if params, ok := s.RawParams.(*T); ok {
		return params, true
	}
	return nil, false
}

func GetParam(s SiteConfigWrapper) (SiteConfig, error) {
	var (
		params SiteConfig
		ok     bool
	)
	switch s.Type {
	case model.TYPE_MASTODON:
		params, ok = GetParams[MastodonConfig](s)
		if !ok {
			return nil, errors.New("cannot fetch MastodonConfig")
		}
	case model.TYPE_MISSKEY:
		params, ok = GetParams[MisskeyConfig](s)
		if !ok {
			return nil, errors.New("cannot fetch MisskeyConfig")
		}
	case model.TYPE_SCRAPBOX:
		var ok bool
		params, ok = GetParams[ScrapboxConfig](s)
		if !ok {
			return nil, errors.New("cannot fetch ScrapboxConfig")
		}
	case model.TYPE_BLOG:
		var ok bool
		params, ok = GetParams[BlogConfig](s)
		if !ok {
			return nil, errors.New("cannot fetch BlogConfig")
		}
	default:
		return nil, fmt.Errorf("unexpected site type: %s", s.Type)
	}
	params.SetId(s.Id)
	params.SetType(s.Type)
	return params, nil
}

type Config struct {
	DB     DbConfig     `yaml:"db"`
	Picker PickerConfig `yaml:"picker"`
	Hb     HbConfig     `yaml:"hb"`
}

type HbConfig struct {
	Url         string    `yaml:"url"`
	TemplateDir string    `yaml:"template_dir"`
	TimeZone    string    `yaml:"timezone"`
	Meta        OgpConfig `yaml:"ogp"`
}

type DbConfig struct {
	MigrationDir string `yaml:"migration_dir"`
	DB           string `yaml:"db"`
}

type PickerConfig struct {
	Sites []SiteConfigWrapper `yaml:"sites"`
}

type SiteConfig interface {
	SetId(id string)
	SetType(typ string)
	GetType() string
	CompleteMetadata(ctx context.Context, id string) error
	GetMetadata() *hb.Site
}

type SiteConfigWrapper struct {
	Type      string      `yaml:"type"`
	Id        string      `yaml:"id"`
	RawParams interface{} `yaml:"params"`
}

type OgpConfig struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}
