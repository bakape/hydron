package fetch

import (
	"io"
	"net/http"
	"strings"
	"time"
)

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
