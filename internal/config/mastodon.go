package config

type MastodonConfig struct {
	Type      string `yaml:"type"`
	Id        string `yaml:"id"`
	SourceUrl string `yaml:"source_url"`
	SiteUrl   string `yaml:"site_url"`
	Name      string `yaml:"name"`
	IconUrl   string `yaml:"icon_url"`
}

func (c *MastodonConfig) GetType() string {
	return c.Type
}
