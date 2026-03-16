package downloader

import (
	"crypto/tls"
	"net/http"
	"time"
)

var client = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        256,
		MaxConnsPerHost:     64,
		MaxIdleConnsPerHost: 64,
		DisableCompression:  true,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   false,
		TLSNextProto:        map[string]func(string, *tls.Conn) http.RoundTripper{},
	},
}
