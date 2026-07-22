//go:build dev

package ui

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func Handler() http.Handler {
	target, _ := url.Parse("http://localhost:3001")
	return httputil.NewSingleHostReverseProxy(target)
}
