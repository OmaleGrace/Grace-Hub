package httpapi

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web
var webFS embed.FS

func staticHandler() http.Handler {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		panic(err)
	}
	return http.FileServerFS(sub)
}