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

func TestDelAppsCmdArgs(t *testing.T) {
	tests := []struct {
		name              string
		managementProject string
		locations         []string
		wantErr           bool
	}{
		{
			name:              "missing management-project",
			managementProject: "",
			locations:         []string{"us-central1"},
			wantErr:           true,
		},
		{
			name:              "missing locations",
			managementProject: "my-project",
			locations:         []string{},
			wantErr:           true,
		},
		{
			name:              "valid arguments",
			managementProject: "my-project",
			locations:         []string{"us-central1"},
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			managementProject = tt.managementProject
			locations = tt.locations
			err := DelAppsCmd.Args(DelAppsCmd, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("DelAppsCmd.Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetDelAppExample(t *testing.T) {
	for i := 0; i < len(delAppsCmdExamples); i++ {
		ex := GetDelAppExample(i)
		if ex == "" {
			t.Errorf("GetDelAppExample(%d) returned empty string", i)
		}
	}
}

func TestDelAppsCmdProperties(t *testing.T) {
	if DelAppsCmd.Use != "delete" {
		t.Errorf("expected Use to be 'delete', got %q", DelAppsCmd.Use)
	}
	if DelAppsCmd.Short != "Delete App Hub Applications" {
		t.Errorf("expected Short to be 'Delete App Hub Applications', got %q", DelAppsCmd.Short)
	}
}
