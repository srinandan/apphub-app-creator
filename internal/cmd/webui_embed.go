// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build embedui

package cmd

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"

	"github.com/gorilla/mux"
)

// webDistFS holds the production build of the Vue frontend. The container image
// populates internal/cmd/webdist from the `npm run build` output before running
// `go build -tags embedui`; the directory is git-ignored and never committed.
//
//go:embed all:webdist
var webDistFS embed.FS

// registerUI serves the embedded single-page app from the root path. Requests
// that don't map to a real asset fall back to index.html so client-side routing
// and deep links resolve.
func registerUI(r *mux.Router) {
	logger := slog.Default()

	sub, err := fs.Sub(webDistFS, "webdist")
	if err != nil {
		logger.Error("Failed to load embedded web UI, falling back to health endpoint", "error", err)
		r.HandleFunc("/", healthHandler).Methods("GET")
		return
	}

	fileServer := http.FileServer(http.FS(sub))
	r.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		asset := strings.TrimPrefix(path.Clean(req.URL.Path), "/")
		if asset == "" {
			asset = "index.html"
		}
		if _, statErr := fs.Stat(sub, asset); statErr != nil {
			// Not a real file: serve the SPA entry point instead of a 404.
			req = req.Clone(req.Context())
			req.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, req)
	}))
}
