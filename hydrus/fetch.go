package hydrus

import (
	"bytes"
	"compress/zlib"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Hydrus PTR creds
const (
	RepoAddr       = "https://hydrus.no-ip.org:45871"
	RepoKey        = "4a285629721ca442541ef2c15ea17d1f7f7578b0c3f4f5f2a05f8f0ab297786f"
	RepoCounterKey = "hydrus_repo_counter:default"
)

// Client is needed, because Hydrus uses a self-signed cert.
//  Must bypass the regular security checks.
var Client = http.Client{
	Transport: &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// GetUpdate fetches and parse a new Hydrus tag repo update
func GetUpdate(id string, session *http.Cookie) (
	com Committer, err error,
) {
	url := RepoAddr + "/update?update_hash=" + id
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.AddCookie(session)
	res, err := Client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	buf, err := decodeZlib(res)
	if err != nil {
		return
	}

	// Initial decode and determine update type
	var arr []json.RawMessage
	err = json.Unmarshal(buf, &arr)
	if err != nil {
		return
	}
	if len(arr) < 3 {
		err = invalidStructureError(buf)
		return
	}

	var typ uint
	err = json.Unmarshal(arr[0], &typ)
	if err != nil {
		return
	}
	switch typ {
	case 36:
		return parseDefinitionUpdate(arr)
	case 34:
		return parseContentUpdate(arr)
	default:
		return nil, fmt.Errorf("unknown update type: %d", typ)
	}
}

// GetRepoMeta fetches a list of new updates
func GetRepoMeta(session *http.Cookie, ctr uint64) (meta RepoMeta, err error) {
	url := RepoAddr + "/metadata?since=" + strconv.FormatUint(ctr, 10)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	req.AddCookie(session)
	res, err := Client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	buf, err := decodeZlib(res)
	if err != nil {
		return
	}
	err = meta.UnmarshalJSON(buf)
	return
}

func decodeZlib(res *http.Response) (buf []byte, err error) {
	var w bytes.Buffer

	zr, err := zlib.NewReader(res.Body)
	if err != nil {
		return
	}
	defer zr.Close()

	_, err = io.Copy(&w, zr)
	buf = w.Bytes()
	return
}

// GetSession retrieves the session key to use for connections
func GetSession() (session *http.Cookie, err error) {
	req, err := http.NewRequest("GET", RepoAddr+"/session_key", nil)
	if err != nil {
		return
	}
	req.Header.Set("Hydrus-Key", RepoKey)
	res, err := Client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	for _, c := range res.Cookies() {
		if c.Name == "session_key" {
			session = c
			break
		}
	}
	if session == nil {
		err = errors.New("no session key received")
	}
	return
}
