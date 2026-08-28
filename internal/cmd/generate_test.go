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
	"internal/clilog"
	"os"
	"path/filepath"
	"testing"

	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/spf13/pflag"
)

func resetFlags() {
	parent = ""
	managementProject = ""
	locations = []string{}
	projectKeys = []string{}
	GenAppsCmd.Flags().Visit(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
	})
}

func TestGenAppsCmdArgs(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		parent            string
		managementProject string
		locations         []string
		projectKeys       []string
		wantErr           bool
	}{
		{
			name:      "missing parent",
			args:      []string{"--label-key", "test"},
			parent:    "",
			locations: []string{"us-central1"},
			wantErr:   true,
		},
		{
			name:      "invalid parent format",
			args:      []string{"--label-key", "test"},
			parent:    "my-invalid-project",
			locations: []string{"us-central1"},
			wantErr:   true,
		},
		{
			name:              "folder parent missing management project",
			args:              []string{"--label-key", "test"},
			parent:            "folders/123456",
			managementProject: "",
			locations:         []string{"us-central1"},
			wantErr:           true,
		},
		{
			name:      "missing locations",
			args:      []string{"--label-key", "test"},
			parent:    "projects/my-project",
			locations: []string{},
			wantErr:   true,
		},
		{
			name:      "label-value without label-key",
			args:      []string{"--label-value", "val"},
			parent:    "projects/my-project",
			locations: []string{"us-central1"},
			wantErr:   true,
		},
		{
			name:      "tag-value without tag-key",
			args:      []string{"--tag-value", "val"},
			parent:    "projects/my-project",
			locations: []string{"us-central1"},
			wantErr:   true,
		},
		{
			name:      "log-label-key without log-label-value",
			args:      []string{"--log-label-key", "key"},
			parent:    "projects/my-project",
			locations: []string{"us-central1"},
			wantErr:   true,
		},
		{
			name:              "folder parent with log-label-key not allowed",
			args:              []string{"--log-label-key", "key", "--log-label-value", "val"},
			parent:            "folders/123456",
			managementProject: "my-mgmt-project",
			locations:         []string{"us-central1"},
			wantErr:           true,
		},
		{
			name:      "invalid app-name starting with uppercase or digit",
			args:      []string{"--label-key", "test", "--app-name", "123App"},
			parent:    "projects/my-project",
			locations: []string{"us-central1"},
			wantErr:   true,
		},
		{
			name:        "multiple project keys with project parent not allowed",
			args:        []string{"--app-name", "myapp"},
			parent:      "projects/my-project",
			locations:   []string{"us-central1"},
			projectKeys: []string{"proj1", "proj2"},
			wantErr:     true,
		},
		{
			name:      "valid args with label-key",
			args:      []string{"--label-key", "test"},
			parent:    "projects/my-project",
			locations: []string{"us-central1"},
			wantErr:   false,
		},
		{
			name:      "valid args with label-key and label-value",
			args:      []string{"--label-key", "test", "--label-value", "val"},
			parent:    "projects/my-project",
			locations: []string{"us-central1"},
			wantErr:   false,
		},
		{
			name:      "valid args with tag-key",
			args:      []string{"--tag-key", "test"},
			parent:    "projects/my-project",
			locations: []string{"us-central1"},
			wantErr:   false,
		},
		{
			name:      "valid args with tag-key and tag-value",
			args:      []string{"--tag-key", "test", "--tag-value", "val"},
			parent:    "projects/my-project",
			locations: []string{"us-central1"},
			wantErr:   false,
		},
		{
			name:      "valid args with contains",
			args:      []string{"--contains", "test"},
			parent:    "projects/my-project",
			locations: []string{"us-central1"},
			wantErr:   false,
		},
		{
			name:      "valid args with log-label",
			args:      []string{"--log-label-key", "key", "--log-label-value", "val"},
			parent:    "projects/my-project",
			locations: []string{"us-central1"},
			wantErr:   false,
		},
		{
			name:              "valid args with multiple project keys under folder parent",
			args:              []string{"--app-name", "myapp"},
			parent:            "folders/123456",
			managementProject: "my-mgmt-project",
			locations:         []string{"us-central1"},
			projectKeys:       []string{"proj1", "proj2"},
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags()
			parent = tt.parent
			managementProject = tt.managementProject
			locations = tt.locations
			projectKeys = tt.projectKeys

			_ = GenAppsCmd.ParseFlags(tt.args)
			err := GenAppsCmd.Args(GenAppsCmd, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("GenAppsCmd.Args() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenAppsCmdRunE(t *testing.T) {
	tempDir := t.TempDir()
	validAttrFile := filepath.Join(tempDir, "attributes.json")
	_ = os.WriteFile(validAttrFile, []byte(`{"criticality":{"type":"MISSION_CRITICAL"}}`), 0644)
	validAssetTypesFile := filepath.Join(tempDir, "asset_types.csv")
	_ = os.WriteFile(validAssetTypesFile, []byte("run.googleapis.com/Service\n"), 0644)

	cleanupClient := client.SetAppHubClientFuncForTest(func() (client.AppHubClient, error) {
		return &testAppHubClient{}, nil
	})
	defer cleanupClient()

	cleanupSearch := client.SetSearchAssetsFuncForTest(func(parent, labelKey, labelValue, tagKey, tagValue, contains string, locations []string, assetTypesData []byte) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{
			{
				Name:      "//run.googleapis.com/projects/p1/locations/us-central1/services/svc1",
				AssetType: "run.googleapis.com/Service",
				Location:  "us-central1",
				Labels:    map[string]string{"env": "prod", "app": "myapp"},
			},
		}, nil
	})
	defer cleanupSearch()

	cleanupSearchK8s := client.SetSearchKubernetesFuncForTest(func(parent string, locations []string) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{
			{
				Name:                   "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/ns1/apps/deployments/dep1",
				AssetType:              "apps.k8s.io/Deployment",
				Location:               "us-central1",
				ParentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/ns1",
			},
		}, nil
	})
	defer cleanupSearchK8s()

	cleanupSearchK8sApps := client.SetSearchKubernetesAppsFuncForTest(func(parent string, locations []string) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{
			{
				Name:      "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/ns1/apps/deployments/dep1",
				AssetType: "apps.k8s.io/Deployment",
				Location:  "us-central1",
				Labels: map[string]string{
					"app.kubernetes.io/name": "my-k8s-app",
				},
			},
		}, nil
	})
	defer cleanupSearchK8sApps()

	tests := []struct {
		name      string
		args      []string
		parent    string
		locations []string
		wantErr   bool
	}{
		{
			name:      "attributes file not found",
			args:      []string{"--label-key", "test", "--attributes", "nonexistent-file.json"},
			parent:    "projects/test",
			locations: []string{"us-central1"},
			wantErr:   true,
		},
		{
			name:      "asset-types file not found",
			args:      []string{"--label-key", "test", "--asset-types", "nonexistent-file.csv"},
			parent:    "projects/test",
			locations: []string{"us-central1"},
			wantErr:   true,
		},
		{
			name:      "invalid parent for GetProjectID when management-project not set",
			args:      []string{"--label-key", "test"},
			parent:    "folders/12345",
			locations: []string{"us-central1"},
			wantErr:   true,
		},
		{
			name:      "successful run with auto-detect and report-only",
			args:      []string{"--auto-detect", "--report-only", "--attributes", validAttrFile},
			parent:    "projects/test",
			locations: []string{"us-central1"},
			wantErr:   false,
		},
		{
			name:      "successful run with per-k8s-namespace and report-only",
			args:      []string{"--per-k8s-namespace", "--report-only"},
			parent:    "projects/test",
			locations: []string{"us-central1"},
			wantErr:   false,
		},
		{
			name:      "successful run with per-k8s-app-label and report-only",
			args:      []string{"--per-k8s-app-label", "--report-only"},
			parent:    "projects/test",
			locations: []string{"us-central1"},
			wantErr:   false,
		},
		{
			name:      "successful run with label-key and asset-types",
			args:      []string{"--label-key", "env", "--report-only", "--asset-types", validAssetTypesFile},
			parent:    "projects/test",
			locations: []string{"us-central1"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clilog.Init(nil)
			resetFlags()
			parent = tt.parent
			locations = tt.locations

			_ = GenAppsCmd.ParseFlags(tt.args)
			err := GenAppsCmd.RunE(GenAppsCmd, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("GenAppsCmd.RunE() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidAppName(t *testing.T) {
	tests := []struct {
		appName string
		want    bool
	}{
		{"myapp", true},
		{"my-app-1", true},
		{"a", true},
		{"MyApp", false},
		{"123app", false},
		{"-myapp", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.appName, func(t *testing.T) {
			got := isValidAppName(tt.appName)
			if got != tt.want {
				t.Errorf("isValidAppName(%q) = %v, want %v", tt.appName, got, tt.want)
			}
		})
	}
}

func TestGetGenAppExample(t *testing.T) {
	for i := 0; i < len(genAppsCmdExamples); i++ {
		ex := GetGenAppExample(i)
		if ex == "" {
			t.Errorf("GetGenAppExample(%d) returned empty string", i)
		}
	}
}
