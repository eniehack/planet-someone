package config

import (
	"fmt"
	"io"
	"log"
	"os"

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

func NewSiteConfig(typ string) (SiteConfig, error) {
	var c SiteConfig
	switch typ {
	case model.TYPE_MASTODON:
		c = new(MastodonConfig)
	case model.TYPE_MISSKEY:
		c = new(MisskeyConfig)
	case model.TYPE_SCRAPBOX:
		c = new(ScrapboxConfig)
	case model.TYPE_BLOG:
		c = new(BlogConfig)
	default:
		return nil, fmt.Errorf("unknown type: %s", typ)
	}
	return c, nil
}

func (sw *SiteConfigWrapper) UnmarshalYAML(value *yaml.Node) error {
	type rawWrapper struct {
		Type string `yaml:"type"`
	}
	var raw rawWrapper
	if err := value.Decode(&raw); err != nil {
		return err
	}
	sw.Type = raw.Type

	// 型に応じて具体的な構造体を割り当てる
	var cfg SiteConfig
	switch raw.Type {
	case model.TYPE_MASTODON:
		cfg = new(MastodonConfig)
	case model.TYPE_MISSKEY:
		cfg = new(MisskeyConfig)
	case model.TYPE_SCRAPBOX:
		cfg = new(ScrapboxConfig)
	case model.TYPE_BLOG:
		cfg = new(BlogConfig)
	default:
		return fmt.Errorf("unknown type: %s", raw.Type)
	}

	// 構造体にデコード
	if err := value.Decode(cfg); err != nil {
		return err
	}
	sw.SiteConfig = cfg
	return nil
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
	GetType() string
}

type SiteConfigWrapper struct {
	Type       string                 `yaml:"type"`
	SiteConfig SiteConfig             `yaml:"-"`
	RawConfig  map[string]interface{} `yaml:",inline"`
}

type OgpConfig struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}
