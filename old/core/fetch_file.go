package core

import (
	"io"
	"net/http"
	"strings"
	"time"
)

// Returns, if URL shoiuld be fetched over network
func IsFetchable(url string) bool {
	for _, pre := range [...]string{
		"http://", "https://", "ftp://", "ftps://",
	} {
		if strings.HasPrefix(url, pre) {
			return true
		}
	}
	return false
}

// Fetch a file over network
func FetchFile(url string) (rc io.ReadCloser, err error) {
	cl := http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	if strings.Contains(url, "pximg.net") {
		req.Header.Set("Referer", "https://www.pixiv.net")
	}

	res, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}
