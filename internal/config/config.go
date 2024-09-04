package config

import (
	"io"
	"log"
	"os"

	"gopkg.in/yaml.v3"
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

type Config struct {
	DB     DbConfig     `yaml:"db"`
	Picker PickerConfig `yaml:"picker"`
	Hb     HbConfig     `yaml:"hb"`
}

type HbConfig struct {
	Url      string    `yaml:"url"`
	TimeZone string    `yaml:"timezone"`
	Meta     OgpConfig `yaml:"ogp"`
}

type DbConfig struct {
	MigrationDir string `yaml:"migration_dir"`
	DB           string `yaml:"db"`
}

type PickerConfig struct {
	Sites []SiteConfig `yaml:"sites"`
}

type SiteConfig struct {
	Id        string `yaml:"id"`
	SourceUrl string `yaml:"source_url"`
	SiteUrl   string `yaml:"site_url"`
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
}

type OgpConfig struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}
