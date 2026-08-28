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
	"internal/client"
	"testing"

	"github.com/spf13/pflag"
)

func TestGetStringParam(t *testing.T) {
	tests := []struct {
		name     string
		flag     *pflag.Flag
		expected string
	}{
		{
			name:     "nil flag",
			flag:     nil,
			expected: "",
		},
		{
			name: "flag with value",
			flag: &pflag.Flag{
				Value: func() pflag.Value {
					var v pflag.Value
					v = new(stringValue)
					v.Set("test")
					return v
				}(),
			},
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := GetStringParam(tt.flag)
			if actual != tt.expected {
				t.Errorf("GetStringParam() = %v, want %v", actual, tt.expected)
			}
		})
	}
}

func TestIsValidResourceFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid project resource",
			input:    "projects/my-project-123",
			expected: true,
		},
		{
			name:     "valid folder resource",
			input:    "folders/123456789",
			expected: true,
		},
		{
			name:     "projects/ prefix with empty ID",
			input:    "projects/",
			expected: false,
		},
		{
			name:     "folders/ prefix with empty ID",
			input:    "folders/",
			expected: false,
		},
		{
			name:     "arbitrary string without valid prefix",
			input:    "my-project-123",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "invalid prefix",
			input:    "organizations/12345",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidResourceFormat(tt.input)
			if got != tt.expected {
				t.Errorf("IsValidResourceFormat(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetProjectID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantID    string
		expectErr bool
	}{
		{
			name:      "valid project resource",
			input:     "projects/my-project-123",
			wantID:    "my-project-123",
			expectErr: false,
		},
		{
			name:      "missing prefix",
			input:     "my-project-123",
			wantID:    "",
			expectErr: true,
		},
		{
			name:      "empty project ID after prefix",
			input:     "projects/",
			wantID:    "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := GetProjectID(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("GetProjectID(%q) error = %v, expectErr %v", tt.input, err, tt.expectErr)
			}
			if gotID != tt.wantID {
				t.Errorf("GetProjectID(%q) = %q, want %q", tt.input, gotID, tt.wantID)
			}
		})
	}
}

func TestIsFolder(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "folder resource",
			input:    "folders/123456",
			expected: true,
		},
		{
			name:     "project resource",
			input:    "projects/my-project",
			expected: false,
		},
		{
			name:     "plain string",
			input:    "123456",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFolder(tt.input)
			if got != tt.expected {
				t.Errorf("IsFolder(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPrintGeneratedApplication(t *testing.T) {
	apps := map[string]client.Application{
		"app1": {
			Name: "app1",
			Workloads: []client.ResourceIdentifier{
				{
					AppHubID: "workload-1",
					URI:      "//compute.googleapis.com/projects/p/zones/z/instances/i",
				},
			},
			Services: []client.ResourceIdentifier{
				{
					AppHubID: "service-1",
					URI:      "//run.googleapis.com/projects/p/locations/l/services/s",
				},
			},
		},
	}

	// Should not panic
	PrintGeneratedApplication(apps)
	PrintGeneratedApplication(nil)
}

type stringValue string

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Type() string {
	return "string"
}

func (s *stringValue) String() string {
	return string(*s)
}
