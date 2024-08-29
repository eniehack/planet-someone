package model

const (
	TYPE_BLOG = iota
	TYPE_SCRAPBOX
	TYPE_MASTODON
	TYPE_MISSKEY
)

func LookupTypeNumber(siteType string) int {
	switch siteType {
	case "blog":
		return TYPE_BLOG
	case "scrapbox":
		return TYPE_SCRAPBOX
	case "misskey":
		return TYPE_MISSKEY
	case "mastodon":
		return TYPE_MASTODON
	default:
		return -1
	}
}
