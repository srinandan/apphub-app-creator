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
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"

	apphubpb "cloud.google.com/go/apphub/apiv1/apphubpb"
	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/protobuf/proto"
)

func TestMain(m *testing.M) {
	// Keep test output clean by discarding logs from the code under test.
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	os.Exit(m.Run())
}

func TestDescribeRegion(t *testing.T) {
	tests := []struct {
		region  string
		want    string
		wantErr bool
	}{
		{"us-central1", "us-central1", false},
		{"europe-west1", "europe-west1", false},
		{"global", "global", false},
		{"us", "global", false},
		{"eu", "global", false},
		{"asia", "global", false},
		{"unknown-region-99", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.region, func(t *testing.T) {
			got, err := describeRegion(tt.region)
			if (err != nil) != tt.wantErr {
				t.Errorf("describeRegion(%q) error = %v, wantErr %v", tt.region, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("describeRegion(%q) = %q, want %q", tt.region, got, tt.want)
			}
		})
	}
}

func TestIsValidAppNameClient(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"myapp", true},
		{"my-app", true},
		{"a", true},
		{"MyApp", false},
		{"123app", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidAppName(tt.name)
			if got != tt.want {
				t.Errorf("isValidAppName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestGetProjectID(t *testing.T) {
	// Plain string returns directly
	plain := getProjectID("my-project-id", context.Background())
	if plain != "my-project-id" {
		t.Errorf("getProjectID(plain) = %q, want %q", plain, "my-project-id")
	}

	// Resource name without valid credentials returns "unknown"
	unknown := getProjectID("projects/123456789", context.Background())
	if unknown != "unknown" {
		t.Errorf("getProjectID(projects/...) = %q, want %q", unknown, "unknown")
	}
}

func TestCreateShortSHA(t *testing.T) {
	s1 := createShortSHA("cluster-1")
	s2 := createShortSHA("cluster-1")
	s3 := createShortSHA("cluster-2")

	if len(s1) != 7 {
		t.Errorf("expected SHA length 7, got %d", len(s1))
	}
	if s1 != s2 {
		t.Errorf("expected deterministic hash: %q != %q", s1, s2)
	}
	if s1 == s3 {
		t.Errorf("expected different hashes for different inputs")
	}
}

func TestGetAppNameForKubernetes(t *testing.T) {
	parentName := "//container.googleapis.com/projects/my-project/locations/us-central1/clusters/my-cluster/k8s/namespaces/my-namespace"
	got := getAppNameForKubernetes(parentName)

	if got == "" {
		t.Errorf("getAppNameForKubernetes returned empty string")
	}
	expectedPrefix := "my-namespace-"
	if len(got) < len(expectedPrefix) || got[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("expected app name to start with %q, got %q", expectedPrefix, got)
	}
}

func TestGetAppName(t *testing.T) {
	asset := &assetpb.ResourceSearchResult{
		Project: "projects/12345",
		Labels: map[string]string{
			"env": "production",
		},
		Tags: []*assetpb.Tag{
			{
				TagKey:   proto.String("tagKeys/123/cost-center"),
				TagValue: proto.String("tagValues/456/finance"),
			},
		},
		EffectiveTags: []*assetpb.EffectiveTagDetails{
			{
				EffectiveTags: []*assetpb.Tag{
					{
						TagKey:   proto.String("tagKeys/123/department"),
						TagValue: proto.String("tagValues/456/engineering"),
					},
				},
			},
		},
	}

	tests := []struct {
		name       string
		labelKey   string
		tagKey     string
		contains   string
		labelValue string
		tagValue   string
		want       string
	}{
		{
			name:       "explicit labelValue",
			labelValue: "custom-app",
			want:       "custom-app",
		},
		{
			name:     "explicit tagValue",
			tagValue: "custom-tag-app",
			want:     "custom-tag-app",
		},
		{
			name:     "match labelKey",
			labelKey: "env",
			want:     "production",
		},
		{
			name:   "match direct tagKey",
			tagKey: "cost-center",
			want:   "finance",
		},
		{
			name:   "match effective tagKey",
			tagKey: "department",
			want:   "engineering",
		},
		{
			name:     "fallback to contains",
			contains: "contains-app",
			want:     "contains-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getAppName(tt.labelKey, tt.tagKey, tt.contains, tt.labelValue, tt.tagValue, asset)
			if got != tt.want {
				t.Errorf("getAppName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetAppNameFromAsset(t *testing.T) {
	tests := []struct {
		name  string
		asset *assetpb.ResourceSearchResult
		want  string
	}{
		{
			name: "label with app in key",
			asset: &assetpb.ResourceSearchResult{
				Project: "my-project",
				Labels: map[string]string{
					"app": "frontend-app",
				},
			},
			want: "frontend-app",
		},
		{
			name: "k8s app label",
			asset: &assetpb.ResourceSearchResult{
				Project: "my-project",
				Labels: map[string]string{
					K8S_APP_LABEL: "k8s-app",
				},
			},
			want: "k8s-app",
		},
		{
			name: "tag with app in key",
			asset: &assetpb.ResourceSearchResult{
				Project: "my-project",
				Tags: []*assetpb.Tag{
					{
						TagKey:   proto.String("tagKeys/123/app-name"),
						TagValue: proto.String("tagValues/456/backend-app"),
					},
				},
			},
			want: "backend-app",
		},
		{
			name: "fallback to plain project string",
			asset: &assetpb.ResourceSearchResult{
				Project: "fallback-project",
			},
			want: "fallback-project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getAppNameFromAsset(tt.asset)
			if got != tt.want {
				t.Errorf("getAppNameFromAsset() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateAppsAssetInventoryMocked(t *testing.T) {
	origSearch := searchAssetsFunc
	origGetClient := getAppHubClientFunc
	defer func() {
		searchAssetsFunc = origSearch
		getAppHubClientFunc = origGetClient
	}()

	mockClient := &mockAppHubClient{
		lookupDiscoveredServiceFunc: func(ctx context.Context, req *apphubpb.LookupDiscoveredServiceRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredServiceResponse, error) {
			return &apphubpb.LookupDiscoveredServiceResponse{
				DiscoveredService: &apphubpb.DiscoveredService{
					Name: "projects/mp/locations/us-central1/discoveredServices/service-1",
				},
			}, nil
		},
	}

	getAppHubClientFunc = func() (appHubClient, error) {
		return mockClient, nil
	}

	searchAssetsFunc = func(parent, labelKey, labelValue, tagKey, tagValue, contains string, locations []string, assetTypesData []byte) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{
			{
				Name:      "//run.googleapis.com/projects/p1/locations/us-central1/services/svc1",
				AssetType: "run.googleapis.com/Service",
				Location:  "us-central1",
				Labels:    map[string]string{"env": "prod"},
			},
		}, nil
	}

	apps, err := GenerateAppsAssetInventory("projects/p1", "mp", "env", "prod", "", "", "", []string{"us-central1"}, nil, nil, true)
	if err != nil {
		t.Fatalf("GenerateAppsAssetInventory failed: %v", err)
	}

	if len(apps) != 1 {
		t.Fatalf("expected 1 application, got %d", len(apps))
	}
	if app, ok := apps["prod"]; !ok {
		t.Errorf("expected application key 'prod', got %v", apps)
	} else if len(app.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(app.Services))
	}
}

func TestGenerateAppsAssetInventoryErrors(t *testing.T) {
	origSearch := searchAssetsFunc
	origGetClient := getAppHubClientFunc
	defer func() {
		searchAssetsFunc = origSearch
		getAppHubClientFunc = origGetClient
	}()

	// Case 1: searchAssets returns error
	searchAssetsFunc = func(parent, labelKey, labelValue, tagKey, tagValue, contains string, locations []string, assetTypesData []byte) ([]*assetpb.ResourceSearchResult, error) {
		return nil, fmt.Errorf("search failed")
	}
	_, err := GenerateAppsAssetInventory("projects/p1", "mp", "env", "prod", "", "", "", []string{"us-central1"}, nil, nil, true)
	if err == nil {
		t.Errorf("expected error when search fails, got nil")
	}

	// Case 2: searchAssets returns empty list
	searchAssetsFunc = func(parent, labelKey, labelValue, tagKey, tagValue, contains string, locations []string, assetTypesData []byte) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{}, nil
	}
	_, err = GenerateAppsAssetInventory("projects/p1", "mp", "env", "prod", "", "", "", []string{"us-central1"}, nil, nil, true)
	if err == nil {
		t.Errorf("expected error when no assets found, got nil")
	}

	// Case 3: getAppHubClient returns error
	searchAssetsFunc = func(parent, labelKey, labelValue, tagKey, tagValue, contains string, locations []string, assetTypesData []byte) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{
			{
				Name:      "//run.googleapis.com/projects/p1/locations/us-central1/services/svc1",
				AssetType: "run.googleapis.com/Service",
				Location:  "us-central1",
			},
		}, nil
	}
	getAppHubClientFunc = func() (appHubClient, error) {
		return nil, fmt.Errorf("client init failed")
	}
	_, err = GenerateAppsAssetInventory("projects/p1", "mp", "env", "prod", "", "", "", []string{"us-central1"}, nil, nil, true)
	if err == nil {
		t.Errorf("expected error when client init fails, got nil")
	}
}

func TestGenerateAppsCloudLogging(t *testing.T) {
	origFilter := filterLogsFunc
	origGetClient := getAppHubClientFunc
	defer func() {
		filterLogsFunc = origFilter
		getAppHubClientFunc = origGetClient
	}()

	mockClient := &mockAppHubClient{
		lookupDiscoveredServiceFunc: func(ctx context.Context, req *apphubpb.LookupDiscoveredServiceRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredServiceResponse, error) {
			return &apphubpb.LookupDiscoveredServiceResponse{
				DiscoveredService: &apphubpb.DiscoveredService{
					Name: "projects/mp/locations/us-central1/discoveredServices/service-1",
				},
			}, nil
		},
	}

	getAppHubClientFunc = func() (appHubClient, error) {
		return mockClient, nil
	}

	filterLogsFunc = func(projectID, labelKey, labelValue string, locations []string) (map[string]logAsset, error) {
		return map[string]logAsset{
			"//run.googleapis.com/projects/p1/locations/us-central1/services/svc1": {
				Name:       "svc1",
				AppHubType: "discoveredService",
				Location:   "us-central1",
			},
		}, nil
	}

	apps, err := GenerateAppsCloudLogging("p1", "mp", "service_name", "my-log-app", []string{"us-central1"}, nil, true)
	if err != nil {
		t.Fatalf("GenerateAppsCloudLogging failed: %v", err)
	}

	if len(apps) != 1 {
		t.Fatalf("expected 1 application, got %d", len(apps))
	}
	if app, ok := apps["my-log-app"]; !ok {
		t.Errorf("expected application 'my-log-app', got %v", apps)
	} else if len(app.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(app.Services))
	}

	// Error path: filterLogs error
	filterLogsFunc = func(projectID, labelKey, labelValue string, locations []string) (map[string]logAsset, error) {
		return nil, fmt.Errorf("filter failed")
	}
	_, err = GenerateAppsCloudLogging("p1", "mp", "service_name", "my-log-app", []string{"us-central1"}, nil, true)
	if err == nil {
		t.Errorf("expected error when filterLogs fails, got nil")
	}

	// Error path: empty logs
	filterLogsFunc = func(projectID, labelKey, labelValue string, locations []string) (map[string]logAsset, error) {
		return map[string]logAsset{}, nil
	}
	_, err = GenerateAppsCloudLogging("p1", "mp", "service_name", "my-log-app", []string{"us-central1"}, nil, true)
	if err == nil {
		t.Errorf("expected error when no logs found, got nil")
	}
}

func TestGenerateAppsPerNamespace(t *testing.T) {
	origSearchK8s := searchKubernetesFunc
	origGetClient := getAppHubClientFunc
	defer func() {
		searchKubernetesFunc = origSearchK8s
		getAppHubClientFunc = origGetClient
	}()

	mockClient := &mockAppHubClient{
		lookupDiscoveredWorkloadFunc: func(ctx context.Context, req *apphubpb.LookupDiscoveredWorkloadRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredWorkloadResponse, error) {
			return &apphubpb.LookupDiscoveredWorkloadResponse{
				DiscoveredWorkload: &apphubpb.DiscoveredWorkload{
					Name: "projects/mp/locations/us-central1/discoveredWorkloads/wl-1",
				},
			}, nil
		},
	}

	getAppHubClientFunc = func() (appHubClient, error) {
		return mockClient, nil
	}

	searchKubernetesFunc = func(parent string, locations []string) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{
			{
				Name:                   "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/ns1/apps/deployments/dep1",
				AssetType:              "apps.k8s.io/Deployment",
				Location:               "us-central1",
				ParentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/ns1",
			},
		}, nil
	}

	apps, err := GenerateAppsPerNamespace("projects/p1", "mp", []string{"us-central1"}, nil, true)
	if err != nil {
		t.Fatalf("GenerateAppsPerNamespace failed: %v", err)
	}

	if len(apps) != 1 {
		t.Fatalf("expected 1 application, got %d", len(apps))
	}

	// Error path: search error
	searchKubernetesFunc = func(parent string, locations []string) ([]*assetpb.ResourceSearchResult, error) {
		return nil, fmt.Errorf("k8s search failed")
	}
	_, err = GenerateAppsPerNamespace("projects/p1", "mp", []string{"us-central1"}, nil, true)
	if err == nil {
		t.Errorf("expected error when searchKubernetes fails, got nil")
	}

	// Error path: no assets
	searchKubernetesFunc = func(parent string, locations []string) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{}, nil
	}
	_, err = GenerateAppsPerNamespace("projects/p1", "mp", []string{"us-central1"}, nil, true)
	if err == nil {
		t.Errorf("expected error when no assets found, got nil")
	}
}

func TestGenerateKubernetesApps(t *testing.T) {
	origSearchK8sApps := searchKubernetesAppsFunc
	origGetClient := getAppHubClientFunc
	defer func() {
		searchKubernetesAppsFunc = origSearchK8sApps
		getAppHubClientFunc = origGetClient
	}()

	mockClient := &mockAppHubClient{
		lookupDiscoveredWorkloadFunc: func(ctx context.Context, req *apphubpb.LookupDiscoveredWorkloadRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredWorkloadResponse, error) {
			return &apphubpb.LookupDiscoveredWorkloadResponse{
				DiscoveredWorkload: &apphubpb.DiscoveredWorkload{
					Name: "projects/mp/locations/us-central1/discoveredWorkloads/wl-1",
				},
			}, nil
		},
	}

	getAppHubClientFunc = func() (appHubClient, error) {
		return mockClient, nil
	}

	searchKubernetesAppsFunc = func(parent string, locations []string) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{
			{
				Name:      "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/ns1/apps/deployments/dep1",
				AssetType: "apps.k8s.io/Deployment",
				Location:  "us-central1",
				Labels: map[string]string{
					K8S_APP_LABEL: "my-k8s-app",
				},
			},
		}, nil
	}

	apps, err := GenerateKubernetesApps("projects/p1", "mp", []string{"us-central1"}, nil, true)
	if err != nil {
		t.Fatalf("GenerateKubernetesApps failed: %v", err)
	}

	if len(apps) != 1 {
		t.Fatalf("expected 1 application, got %d", len(apps))
	}
	if _, ok := apps["my-k8s-app"]; !ok {
		t.Errorf("expected app 'my-k8s-app', got %v", apps)
	}

	// Error path
	searchKubernetesAppsFunc = func(parent string, locations []string) ([]*assetpb.ResourceSearchResult, error) {
		return nil, fmt.Errorf("search failed")
	}
	_, err = GenerateKubernetesApps("projects/p1", "mp", []string{"us-central1"}, nil, true)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestGenerateFromAll(t *testing.T) {
	origSearch := searchAssetsFunc
	origSearchK8s := searchKubernetesFunc
	origGetClient := getAppHubClientFunc
	defer func() {
		searchAssetsFunc = origSearch
		searchKubernetesFunc = origSearchK8s
		getAppHubClientFunc = origGetClient
	}()

	mockClient := &mockAppHubClient{
		lookupDiscoveredServiceFunc: func(ctx context.Context, req *apphubpb.LookupDiscoveredServiceRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredServiceResponse, error) {
			return &apphubpb.LookupDiscoveredServiceResponse{
				DiscoveredService: &apphubpb.DiscoveredService{
					Name: "projects/mp/locations/us-central1/discoveredServices/service-1",
				},
			}, nil
		},
	}

	getAppHubClientFunc = func() (appHubClient, error) {
		return mockClient, nil
	}

	searchAssetsFunc = func(parent, labelKey, labelValue, tagKey, tagValue, contains string, locations []string, assetTypesData []byte) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{
			{
				Name:      "//run.googleapis.com/projects/p1/locations/us-central1/services/svc1",
				AssetType: "run.googleapis.com/Service",
				Location:  "us-central1",
				Labels:    map[string]string{"app": "autodetected-app"},
			},
		}, nil
	}

	searchKubernetesFunc = func(parent string, locations []string) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{}, nil
	}

	apps, err := GenerateFromAll("projects/p1", "mp", []string{"us-central1"}, nil, true)
	if err != nil {
		t.Fatalf("GenerateFromAll failed: %v", err)
	}

	if len(apps) != 1 {
		t.Fatalf("expected 1 application, got %d", len(apps))
	}
}

func TestGenerateFromProject(t *testing.T) {
	origSearchProj := searchProjectFunc
	origGetClient := getAppHubClientFunc
	defer func() {
		searchProjectFunc = origSearchProj
		getAppHubClientFunc = origGetClient
	}()

	mockClient := &mockAppHubClient{
		lookupDiscoveredServiceFunc: func(ctx context.Context, req *apphubpb.LookupDiscoveredServiceRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredServiceResponse, error) {
			return &apphubpb.LookupDiscoveredServiceResponse{
				DiscoveredService: &apphubpb.DiscoveredService{
					Name: "projects/mp/locations/us-central1/discoveredServices/service-1",
				},
			}, nil
		},
	}

	getAppHubClientFunc = func() (appHubClient, error) {
		return mockClient, nil
	}

	searchProjectFunc = func(parent string, projectIds, locations []string, assetTypesData []byte) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{
			{
				Name:      "//run.googleapis.com/projects/p1/locations/us-central1/services/svc1",
				AssetType: "run.googleapis.com/Service",
				Location:  "us-central1",
			},
		}, nil
	}

	apps, err := GenerateFromProject("folders/123", "mp", "my-project-app", []string{"p1", "p2"}, []string{"us-central1"}, nil, nil, true)
	if err != nil {
		t.Fatalf("GenerateFromProject failed: %v", err)
	}

	if len(apps) != 1 {
		t.Fatalf("expected 1 application, got %d", len(apps))
	}
	if app, ok := apps["my-project-app"]; !ok {
		t.Errorf("expected application 'my-project-app', got %v", apps)
	} else if len(app.Services) != 1 {
		t.Errorf("expected 1 service, got %d", len(app.Services))
	}
}

func TestDeleteApp(t *testing.T) {
	origDeleteApp := deleteAppFunc
	origGetClient := getAppHubClientFunc
	defer func() {
		deleteAppFunc = origDeleteApp
		getAppHubClientFunc = origGetClient
	}()

	getAppHubClientFunc = func() (appHubClient, error) {
		return &mockAppHubClient{}, nil
	}

	deleteAppFunc = func(apiclient appHubClient, projectID, location, appID string) error {
		return nil
	}

	err := DeleteApp("my-mgmt-project", "my-app", []string{"us-central1", "us-east1"})
	if err != nil {
		t.Fatalf("DeleteApp failed: %v", err)
	}

	// Error path
	deleteAppFunc = func(apiclient appHubClient, projectID, location, appID string) error {
		return fmt.Errorf("delete failed")
	}
	err = DeleteApp("my-mgmt-project", "my-app", []string{"us-central1"})
	if err == nil {
		t.Errorf("expected error when deleteAppFunc fails, got nil")
	}
}

func TestGenerateAppsExcludesSystemNamespaces(t *testing.T) {
	origSearchK8s := searchKubernetesFunc
	origGetClient := getAppHubClientFunc
	defer func() {
		searchKubernetesFunc = origSearchK8s
		getAppHubClientFunc = origGetClient
	}()

	mockClient := &mockAppHubClient{
		lookupDiscoveredWorkloadFunc: func(ctx context.Context, req *apphubpb.LookupDiscoveredWorkloadRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredWorkloadResponse, error) {
			return &apphubpb.LookupDiscoveredWorkloadResponse{
				DiscoveredWorkload: &apphubpb.DiscoveredWorkload{
					Name: "projects/mp/locations/us-central1/discoveredWorkloads/wl-user",
				},
			}, nil
		},
	}

	getAppHubClientFunc = func() (appHubClient, error) {
		return mockClient, nil
	}

	// Mix of a system namespace asset (kube-system) and a user namespace asset (default)
	searchKubernetesFunc = func(parent string, locations []string) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{
			{
				Name:                   "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/kube-system/apps/deployments/kube-dns",
				AssetType:              "apps.k8s.io/Deployment",
				Location:               "us-central1",
				ParentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/kube-system",
			},
			{
				Name:                   "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/user-ns/apps/deployments/user-app",
				AssetType:              "apps.k8s.io/Deployment",
				Location:               "us-central1",
				ParentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/user-ns",
			},
		}, nil
	}

	apps, err := GenerateAppsPerNamespace("projects/p1", "mp", []string{"us-central1"}, nil, true)
	if err != nil {
		t.Fatalf("GenerateAppsPerNamespace failed: %v", err)
	}

	// Only 1 app should be generated (user-ns), kube-system should be skipped
	if len(apps) != 1 {
		t.Fatalf("expected 1 application (excluding kube-system), got %d: %v", len(apps), apps)
	}

	for appName := range apps {
		if strings.HasPrefix(appName, "kube-system") {
			t.Errorf("expected kube-system to be excluded, but found app %q", appName)
		}
	}
}

func TestDeduplicateAssets(t *testing.T) {
	assets := []*assetpb.ResourceSearchResult{
		{Name: "//res1"},
		{Name: "//res2"},
		{Name: "//res1"},
		nil,
		{Name: ""},
		{Name: "//res2"},
		{Name: "//res3"},
	}

	deduped := deduplicateAssets(assets)
	if len(deduped) != 3 {
		t.Fatalf("expected 3 unique assets, got %d", len(deduped))
	}
	if deduped[0].Name != "//res1" || deduped[1].Name != "//res2" || deduped[2].Name != "//res3" {
		t.Errorf("unexpected deduped order/contents: %v", deduped)
	}
}

func TestContainsResource(t *testing.T) {
	list := []ResourceIdentifier{
		{AppHubID: "id1", URI: "//uri1"},
		{AppHubID: "id2", URI: "//uri2"},
	}

	if !containsResource(list, "//uri1") {
		t.Errorf("expected //uri1 to be found")
	}
	if containsResource(list, "//uri3") {
		t.Errorf("expected //uri3 not to be found")
	}
}

func TestGenerateFromAllDeduplication(t *testing.T) {
	origSearchAssets := searchAssetsFunc
	origSearchK8s := searchKubernetesFunc
	origGetClient := getAppHubClientFunc
	defer func() {
		searchAssetsFunc = origSearchAssets
		searchKubernetesFunc = origSearchK8s
		getAppHubClientFunc = origGetClient
	}()

	mockClient := &mockAppHubClient{
		lookupDiscoveredWorkloadFunc: func(ctx context.Context, req *apphubpb.LookupDiscoveredWorkloadRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredWorkloadResponse, error) {
			return &apphubpb.LookupDiscoveredWorkloadResponse{
				DiscoveredWorkload: &apphubpb.DiscoveredWorkload{
					Name: "projects/mp/locations/us-central1/discoveredWorkloads/wl-app",
				},
			}, nil
		},
	}

	getAppHubClientFunc = func() (appHubClient, error) {
		return mockClient, nil
	}

	sharedAsset := &assetpb.ResourceSearchResult{
		Name:      "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/ns/apps/deployments/app1",
		AssetType: "apps.k8s.io/Deployment",
		Location:  "us-central1",
		Labels: map[string]string{
			"app": "my-app",
		},
	}

	// Labeled search and k8s search return the SAME asset
	searchAssetsFunc = func(parent, labelKey, labelValue, tagKey, tagValue, contains string, locations []string, assetTypesData []byte) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{sharedAsset}, nil
	}
	searchKubernetesFunc = func(parent string, locations []string) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{sharedAsset}, nil
	}

	apps, err := GenerateFromAll("projects/p1", "mp", []string{"us-central1"}, nil, true)
	if err != nil {
		t.Fatalf("GenerateFromAll failed: %v", err)
	}

	app, exists := apps["my-app"]
	if !exists {
		t.Fatalf("expected app 'my-app' to exist")
	}

	// Should only have 1 workload, not duplicated
	if len(app.Workloads) != 1 {
		t.Fatalf("expected 1 workload after deduplication, got %d", len(app.Workloads))
	}
}

