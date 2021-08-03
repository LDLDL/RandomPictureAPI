package main

type ConfAuth struct {
	Key string `toml:"key"`
}

type ConfSite struct {
	Cert   string `toml:"cert"`
	Key    string `toml:"key"`
	Listen string `toml:"listen"`
}

type Config struct {
	Auth ConfAuth `toml:"auth"`
	Site ConfSite `toml:"site"`
}

func (c *Config) SetDefault() {
	c.Auth.Key = "defaultkey"

	c.Site.Listen = ":8192"
}
