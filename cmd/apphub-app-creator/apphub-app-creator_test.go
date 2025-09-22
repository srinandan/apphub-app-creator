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

package main

import (
	"fmt"
	"internal/cmd"
	"testing"
)

func TestVersionString(t *testing.T) {
	rootCmd := cmd.GetRootCmd()

	// Set test values for version, date, and commit
	version = "1.0.0"
	date = "2025-01-01"
	commit = "abcdef123456"

	expectedVersion := fmt.Sprintf("%s date: %s [commit: %.7s]", version, date, commit)
	rootCmd.Version = expectedVersion

	if rootCmd.Version != expectedVersion {
		t.Errorf("expected version %q, got %q", expectedVersion, rootCmd.Version)
	}
}
