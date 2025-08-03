package hb

import "time"

type Post struct {
	Id         string
	Content    string
	Url        string
	Date       int64
	ParsedDate *time.Time
	Src        string
}

type PageMeta struct {
	Url         string
	Description string
	Title       string
}

type Config struct {
	Meta PageMeta
}

type BinMeta struct {
	Version string
}

type Site struct {
	Url     string
	IconUrl string
	Title   string
}
