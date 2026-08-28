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

package client

import (
	"strings"
	"testing"

	"cloud.google.com/go/logging"
	"google.golang.org/genproto/googleapis/api/monitoredres"
)

func TestGenerateLocationFilter(t *testing.T) {
	tests := []struct {
		name      string
		locations []string
		want      string
	}{
		{
			name:      "empty locations",
			locations: []string{},
			want:      "",
		},
		{
			name:      "single location",
			locations: []string{"us-central1"},
			want:      `(resource.labels.location="us-central1")`,
		},
		{
			name:      "multiple locations",
			locations: []string{"us-central1", "us-east1"},
			want:      `(resource.labels.location="us-central1" OR resource.labels.location="us-east1")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateLocationFilter(tt.locations)
			if got != tt.want {
				t.Errorf("generateLocationFilter() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateResourceTypeFilter(t *testing.T) {
	filter := generateResourceTypeFilter()
	if filter == "" {
		t.Errorf("generateResourceTypeFilter() returned empty string")
	}
	if !strings.Contains(filter, "cloud_run_revision") {
		t.Errorf("expected filter to contain cloud_run_revision, got %q", filter)
	}
	if !strings.Contains(filter, "k8s_pod") {
		t.Errorf("expected filter to contain k8s_pod, got %q", filter)
	}
	if !strings.Contains(filter, "gce_instance_group") {
		t.Errorf("expected filter to contain gce_instance_group, got %q", filter)
	}
}

func TestGetAsset(t *testing.T) {
	tests := []struct {
		name          string
		entry         *logging.Entry
		wantURI       string
		wantAssetName string
		wantType      string
		wantLocation  string
	}{
		{
			name: "cloud_run_revision entry",
			entry: &logging.Entry{
				Resource: &monitoredres.MonitoredResource{
					Type: "cloud_run_revision",
					Labels: map[string]string{
						"project_id":   "my-proj",
						"location":     "us-central1",
						"service_name": "my-service",
					},
				},
			},
			wantURI:       "//run.googleapis.com/projects/my-proj/locations/us-central1/services/my-service",
			wantAssetName: "my-service",
			wantType:      "discoveredService",
			wantLocation:  "us-central1",
		},
		{
			name: "unsupported resource type",
			entry: &logging.Entry{
				Resource: &monitoredres.MonitoredResource{
					Type: "unsupported_type",
				},
			},
			wantURI:       "",
			wantAssetName: "",
			wantType:      "",
			wantLocation:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uri, asset := getAsset(tt.entry)
			if uri != tt.wantURI {
				t.Errorf("getAsset() uri = %q, want %q", uri, tt.wantURI)
			}
			if asset.Name != tt.wantAssetName {
				t.Errorf("getAsset() asset.Name = %q, want %q", asset.Name, tt.wantAssetName)
			}
			if asset.AppHubType != tt.wantType {
				t.Errorf("getAsset() asset.AppHubType = %q, want %q", asset.AppHubType, tt.wantType)
			}
			if asset.Location != tt.wantLocation {
				t.Errorf("getAsset() asset.Location = %q, want %q", asset.Location, tt.wantLocation)
			}
		})
	}
}
