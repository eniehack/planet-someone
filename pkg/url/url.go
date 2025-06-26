package url

import "net/url"

func ResolveAbsUrl(baseUrl *url.URL, path string) (*url.URL, error) {
	relUrl, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	abs := baseUrl.ResolveReference(relUrl)
	return abs, nil
}
