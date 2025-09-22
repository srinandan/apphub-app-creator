
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
	"internal/clilog"
	"testing"

	"github.com/spf13/pflag"
)

func TestGenAppsCmdArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		project string
		region  string
		wantErr bool
	}{
		{
			name:    "missing project",
			args:    []string{"--region", "us-central1", "--label-key", "test"},
			project: "",
			region:  "us-central1",
			wantErr: true,
		},
		{
			name:    "missing region",
			args:    []string{"--project", "test-project", "--label-key", "test"},
			project: "test-project",
			region:  "",
			wantErr: true,
		},
		{
			name:    "missing label-key, tag-key, or contains",
			args:    []string{"--project", "test-project", "--region", "us-central1"},
			project: "test-project",
			region:  "us-central1",
			wantErr: true,
		},
		{
			name:    "valid args with label-key",
			args:    []string{"--project", "test-project", "--region", "us-central1", "--label-key", "test"},
			project: "test-project",
			region:  "us-central1",
			wantErr: false,
		},
		{
			name:    "valid args with tag-key",
			args:    []string{"--project", "test-project", "--region", "us-central1", "--tag-key", "test"},
			project: "test-project",
			region:  "us-central1",
			wantErr: false,
		},
		{
			name:    "valid args with contains",
			args:    []string{"--project", "test-project", "--region", "us-central1", "--contains", "test"},
			project: "test-project",
			region:  "us-central1",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project = tt.project
			region = tt.region
			GenAppsCmd.Flags().Visit(func(f *pflag.Flag) {
				f.Value.Set(f.DefValue)
			})
			GenAppsCmd.ParseFlags(tt.args)
			err := GenAppsCmd.Args(GenAppsCmd, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("GenAppsCmd.Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenAppsCmdRunE(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		project string
		region  string
		wantErr bool
	}{
		{
			name:    "attributes file not found",
			args:    []string{"--project", "test", "--region", "test", "--label-key", "test", "--attributes", "nonexistent"},
			project: "test",
			region:  "test",
			wantErr: true,
		},
		{
			name:    "asset-types file not found",
			args:    []string{"--project", "test", "--region", "test", "--label-key", "test", "--asset-types", "nonexistent"},
			project: "test",
			region:  "test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clilog.Init()
			project = tt.project
			region = tt.region
			GenAppsCmd.ParseFlags(tt.args)
			err := GenAppsCmd.RunE(GenAppsCmd, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("GenAppsCmd.RunE() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
