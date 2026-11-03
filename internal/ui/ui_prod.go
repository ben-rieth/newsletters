//go:build !dev

package ui

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/ben-rieth/newsletter-api/web"
)

func Handler() http.Handler {
	dist, _ := fs.Sub(web.Files, "dist")
	fsHandler := http.FileServerFS(dist)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := web.Files.Open("dist" + r.URL.Path)
		if err != nil {
			w.Header().Set("Cache-Control", "no-cache")
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			fsHandler.ServeHTTP(w, r2)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		fsHandler.ServeHTTP(w, r)
	})
}
