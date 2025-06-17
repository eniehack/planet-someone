package config

type BlogConfig struct {
	Type      string `yaml:"type"`
	Id        string `yaml:"id"`
	SourceUrl string `yaml:"source_url"`
	SiteUrl   string `yaml:"site_url"`
	Name      string `yaml:"name"`
	IconUrl   string `yaml:"icon_url"`
}

func (c *BlogConfig) GetType() string {
	return c.Type
}
