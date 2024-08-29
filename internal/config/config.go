package config

import (
	"io"

	"gopkg.in/yaml.v3"
)

func New(f io.Reader) (*Config, error) {
	config := new(Config)
	if err := yaml.NewDecoder(f).Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}

type Config struct {
	DB     DbConfig     `yaml:"db"`
	Picker PickerConfig `yaml:"picker"`
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
