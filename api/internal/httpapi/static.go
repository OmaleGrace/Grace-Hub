package httpapi

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web
var webFS embed.FS

func webRoot() fs.FS {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		panic(err)
	}
	return sub
}

func staticHandler() http.Handler {
	return http.FileServerFS(webRoot())
}

func pageHandler(name string) http.HandlerFunc {
	root := webRoot()
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, root, name)
	}
}