package config

type MisskeyConfig struct {
	Type        string `yaml:"type"`
	Id          string `yaml:"id"`
	InstanceUrl string `yaml:"instance_url"`
	UserId      string `yaml:"user_id"`
	SourceUrl   string `yaml:"source_url"`
	SiteUrl     string `yaml:"site_url"`
	Name        string `yaml:"name"`
	IconUrl     string `yaml:"icon_url"`
}

func (c *MisskeyConfig) GetType() string {
	return c.Type
}
