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

package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	healthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("healthHandler returned status %d, want %d", rr.Code, http.StatusOK)
	}

	body := strings.TrimSpace(rr.Body.String())
	if body != "OK" {
		t.Errorf("healthHandler returned body %q, want %q", body, "OK")
	}
}

func TestServerCmdProperties(t *testing.T) {
	if ServerCmd.Use != "server" {
		t.Errorf("expected Use to be 'server', got %q", ServerCmd.Use)
	}
	if ServerCmd.Short != "Start an HTTP server" {
		t.Errorf("expected Short to be 'Start an HTTP server', got %q", ServerCmd.Short)
	}
}
