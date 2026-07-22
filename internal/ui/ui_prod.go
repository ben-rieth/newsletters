//go:build !dev

package ui

import (
	"io/fs"
	"net/http"

	"github.com/ben-rieth/newsletter-api/web"
)

func Handler() http.Handler {
	dist, _ := fs.Sub(web.Files, "dist")
	fsHandler := http.FileServerFS(dist)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := web.Files.Open("dist" + r.URL.Path)
		if err != nil {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/"
			fsHandler.ServeHTTP(w, r2)
			return
		}
		fsHandler.ServeHTTP(w, r)
	})
}
