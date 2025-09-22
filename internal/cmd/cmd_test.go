
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
	"testing"
)

func TestCmd(t *testing.T) {
	if Cmd.Use != "apps" {
		t.Errorf("expected Use to be 'apps', got '%s'", Cmd.Use)
	}

	if len(Cmd.Aliases) != 1 || Cmd.Aliases[0] != "applications" {
		t.Errorf("expected Aliases to be ['applications'], got '%s'", Cmd.Aliases)
	}

	if Cmd.Short != "Manage App Hub Applications" {
		t.Errorf("expected Short to be 'Manage App Hub Applications', got '%s'", Cmd.Short)
	}
}
