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

//go:build !embedui

package cmd

import "github.com/gorilla/mux"

// registerUI wires the root path when the web UI is NOT compiled into the
// binary. The frontend assets are only embedded with the "embedui" build tag
// (used by the container image); without it the root path stays a simple
// liveness endpoint so plain `go build` and API-only deployments keep working.
func registerUI(r *mux.Router) {
	r.HandleFunc("/", healthHandler).Methods("GET")
}
