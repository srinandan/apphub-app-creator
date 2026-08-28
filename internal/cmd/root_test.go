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

func TestExecute(t *testing.T) {
	if RootCmd.Short != "Utility to generate App Hub Applications." {
		t.Errorf("expected %q, got %q", "Utility to generate App Hub Applications.", RootCmd.Short)
	}
}

func TestGetRootCmd(t *testing.T) {
	rootCmd := GetRootCmd()
	if rootCmd == nil {
		t.Error("expected rootCmd not to be nil")
	}

	if rootCmd.Use != "apphub-app-creator" {
		t.Errorf("expected Use to be 'apphub-app-creator', got '%s'", rootCmd.Use)
	}
}

func TestRootCmdPersistentPreRunE(t *testing.T) {
	tests := []struct {
		name         string
		level        string
		disableCheck bool
		wantErr      bool
	}{
		{
			name:         "info level with disable check",
			level:        "info",
			disableCheck: true,
			wantErr:      false,
		},
		{
			name:         "warn level",
			level:        "warn",
			disableCheck: true,
			wantErr:      false,
		},
		{
			name:         "error level",
			level:        "error",
			disableCheck: true,
			wantErr:      false,
		},
		{
			name:         "off level",
			level:        "off",
			disableCheck: true,
			wantErr:      false,
		},
		{
			name:         "invalid level",
			level:        "debug_invalid",
			disableCheck: true,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logLevel = tt.level
			disableCheck = tt.disableCheck
			err := RootCmd.PersistentPreRunE(RootCmd, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("RootCmd.PersistentPreRunE() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
