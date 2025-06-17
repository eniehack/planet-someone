package config

type ScrapboxConfig struct {
	Type      string `yaml:"type"`
	Id        string `yaml:"id"`
	SourceUrl string `yaml:"source_url"`
	SiteUrl   string `yaml:"site_url"`
	Name      string `yaml:"name"`
	IconUrl   string `yaml:"icon_url"`
}

func (c *ScrapboxConfig) GetType() string {
	return c.Type
}
