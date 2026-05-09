package http

import (
	"io/fs"
	"net/http"
)

// StaticFileServer wraps an embed.FS to serve static files
type StaticFileServer struct {
	fs fs.FS
}

func NewStaticFileServer(embeddedFS fs.FS) *StaticFileServer {
	return &StaticFileServer{fs: embeddedFS}
}

func (s *StaticFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.FS(s.fs)).ServeHTTP(w, r)
}