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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"internal/client"
	"net/http"
	"net/http/httptest"
	"testing"

	apphub "cloud.google.com/go/apphub/apiv1"
	apphubpb "cloud.google.com/go/apphub/apiv1/apphubpb"
	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/googleapis/gax-go/v2"
)

type testAppHubClient struct{}

func (m *testAppHubClient) LookupDiscoveredService(ctx context.Context, req *apphubpb.LookupDiscoveredServiceRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredServiceResponse, error) {
	return &apphubpb.LookupDiscoveredServiceResponse{
		DiscoveredService: &apphubpb.DiscoveredService{
			Name: "projects/p1/locations/us-central1/discoveredServices/svc-1",
		},
	}, nil
}

func (m *testAppHubClient) LookupDiscoveredWorkload(ctx context.Context, req *apphubpb.LookupDiscoveredWorkloadRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredWorkloadResponse, error) {
	return &apphubpb.LookupDiscoveredWorkloadResponse{
		DiscoveredWorkload: &apphubpb.DiscoveredWorkload{
			Name: "projects/p1/locations/us-central1/discoveredWorkloads/wl-1",
		},
	}, nil
}

func (m *testAppHubClient) GetApplication(ctx context.Context, req *apphubpb.GetApplicationRequest, opts ...gax.CallOption) (*apphubpb.Application, error) {
	return &apphubpb.Application{}, nil
}

func (m *testAppHubClient) CreateApplication(ctx context.Context, req *apphubpb.CreateApplicationRequest, opts ...gax.CallOption) (*apphub.CreateApplicationOperation, error) {
	return nil, nil
}

func (m *testAppHubClient) ListApplications(ctx context.Context, req *apphubpb.ListApplicationsRequest, opts ...gax.CallOption) *apphub.ApplicationIterator {
	return nil
}

func (m *testAppHubClient) CreateService(ctx context.Context, req *apphubpb.CreateServiceRequest, opts ...gax.CallOption) (*apphub.CreateServiceOperation, error) {
	return nil, nil
}

func (m *testAppHubClient) CreateWorkload(ctx context.Context, req *apphubpb.CreateWorkloadRequest, opts ...gax.CallOption) (*apphub.CreateWorkloadOperation, error) {
	return nil, nil
}

func (m *testAppHubClient) ListServices(ctx context.Context, req *apphubpb.ListServicesRequest, opts ...gax.CallOption) *apphub.ServiceIterator {
	return nil
}

func (m *testAppHubClient) ListWorkloads(ctx context.Context, req *apphubpb.ListWorkloadsRequest, opts ...gax.CallOption) *apphub.WorkloadIterator {
	return nil
}

func (m *testAppHubClient) ListDiscoveredServices(ctx context.Context, req *apphubpb.ListDiscoveredServicesRequest, opts ...gax.CallOption) *apphub.DiscoveredServiceIterator {
	return nil
}

func (m *testAppHubClient) ListDiscoveredWorkloads(ctx context.Context, req *apphubpb.ListDiscoveredWorkloadsRequest, opts ...gax.CallOption) *apphub.DiscoveredWorkloadIterator {
	return nil
}

func (m *testAppHubClient) DeleteService(ctx context.Context, req *apphubpb.DeleteServiceRequest, opts ...gax.CallOption) (*apphub.DeleteServiceOperation, error) {
	return nil, nil
}

func (m *testAppHubClient) DeleteWorkload(ctx context.Context, req *apphubpb.DeleteWorkloadRequest, opts ...gax.CallOption) (*apphub.DeleteWorkloadOperation, error) {
	return nil, nil
}

func (m *testAppHubClient) DeleteApplication(ctx context.Context, req *apphubpb.DeleteApplicationRequest, opts ...gax.CallOption) (*apphub.DeleteApplicationOperation, error) {
	return nil, nil
}

func (m *testAppHubClient) Close() error {
	return nil
}

func TestSelectorValidateExclusive(t *testing.T) {
	containsStr := "frontend"
	tests := []struct {
		name     string
		selector Selector
		wantErr  bool
	}{
		{
			name:     "no selector set",
			selector: Selector{},
			wantErr:  true,
		},
		{
			name: "multiple bool flags set",
			selector: Selector{
				AutoDetect:      true,
				PerK8sNamespace: true,
			},
			wantErr: true,
		},
		{
			name: "valid autoDetect",
			selector: Selector{
				AutoDetect: true,
			},
			wantErr: false,
		},
		{
			name: "valid perK8sNamespace",
			selector: Selector{
				PerK8sNamespace: true,
			},
			wantErr: false,
		},
		{
			name: "valid perK8sAppLabel",
			selector: Selector{
				PerK8sAppLabel: true,
			},
			wantErr: false,
		},
		{
			name: "valid label",
			selector: Selector{
				Label: &KeyValue{Key: "env", Value: "prod"},
			},
			wantErr: false,
		},
		{
			name: "label with empty key",
			selector: Selector{
				Label: &KeyValue{Key: "", Value: "prod"},
			},
			wantErr: true,
		},
		{
			name: "valid tag",
			selector: Selector{
				Tag: &KeyValue{Key: "cost-center", Value: "123"},
			},
			wantErr: false,
		},
		{
			name: "tag with empty key",
			selector: Selector{
				Tag: &KeyValue{Key: "", Value: "123"},
			},
			wantErr: true,
		},
		{
			name: "valid logLabel",
			selector: Selector{
				LogLabel: &KeyValue{Key: "service_name", Value: "auth"},
			},
			wantErr: false,
		},
		{
			name: "logLabel with empty key",
			selector: Selector{
				LogLabel: &KeyValue{Key: "", Value: "auth"},
			},
			wantErr: true,
		},
		{
			name: "valid contains",
			selector: Selector{
				Contains: &containsStr,
			},
			wantErr: false,
		},
		{
			name: "valid projectKeys",
			selector: Selector{
				ProjectKeys: &ProjectKey{
					AppName:    "my-app",
					ProjectIds: []string{"proj-1", "proj-2"},
				},
			},
			wantErr: false,
		},
		{
			name: "projectKeys with empty projectIds",
			selector: Selector{
				ProjectKeys: &ProjectKey{
					AppName:    "my-app",
					ProjectIds: []string{},
				},
			},
			wantErr: true,
		},
		{
			name: "projectKeys with empty appName",
			selector: Selector{
				ProjectKeys: &ProjectKey{
					AppName:    "",
					ProjectIds: []string{"proj-1"},
				},
			},
			wantErr: true,
		},
		{
			name: "multiple pointer selectors set",
			selector: Selector{
				Label: &KeyValue{Key: "env", Value: "prod"},
				Tag:   &KeyValue{Key: "tag-1", Value: "val"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.selector.ValidateExclusive()
			if (err != nil) != tt.wantErr {
				t.Errorf("Selector.ValidateExclusive() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScopeValidate(t *testing.T) {
	tests := []struct {
		name    string
		scope   Scope
		wantErr bool
	}{
		{
			name: "valid scope",
			scope: Scope{
				Parent:            "projects/p1",
				ManagementProject: "p1",
				Locations:         []string{"us-central1"},
			},
			wantErr: false,
		},
		{
			name: "missing parent",
			scope: Scope{
				Parent:            "",
				ManagementProject: "p1",
				Locations:         []string{"us-central1"},
			},
			wantErr: true,
		},
		{
			name: "missing management project",
			scope: Scope{
				Parent:            "projects/p1",
				ManagementProject: "",
				Locations:         []string{"us-central1"},
			},
			wantErr: true,
		},
		{
			name: "empty locations",
			scope: Scope{
				Parent:            "projects/p1",
				ManagementProject: "p1",
				Locations:         []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scope.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Scope.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateRequestValidate(t *testing.T) {
	req := GenerateRequest{
		Selector: Selector{AutoDetect: true},
		Scope: Scope{
			Parent:            "projects/p1",
			ManagementProject: "p1",
			Locations:         []string{"us-central1"},
		},
	}

	if err := req.Validate(); err != nil {
		t.Errorf("GenerateRequest.Validate() unexpected error: %v", err)
	}

	invalidReq := GenerateRequest{
		Selector: Selector{},
		Scope:    Scope{},
	}
	if err := invalidReq.Validate(); err == nil {
		t.Errorf("GenerateRequest.Validate() expected error for empty request, got nil")
	}
}

func TestGetString(t *testing.T) {
	val := "hello"
	if got := getString(&val); got != "hello" {
		t.Errorf("getString(&%q) = %q, want %q", val, got, "hello")
	}
	if got := getString(nil); got != "" {
		t.Errorf("getString(nil) = %q, want %q", got, "")
	}
}

func TestGenerateHandlerInvalidPayload(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewBufferString("invalid json"))
	rr := httptest.NewRecorder()

	generateHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("generateHandler returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}
	if errResp.Error == "" {
		t.Errorf("expected non-empty error message in response")
	}
}

func TestGenerateHandlerValidationFailure(t *testing.T) {
	invalidPayload := GenerateRequest{
		Selector: Selector{},
		Scope: Scope{
			Parent:            "projects/p1",
			ManagementProject: "p1",
			Locations:         []string{"us-central1"},
		},
	}
	body, _ := json.Marshal(invalidPayload)

	req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	generateHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("generateHandler returned status %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestGenerateHandlerSuccessSelectors(t *testing.T) {
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

	cleanupSearchProj := client.SetSearchProjectFuncForTest(func(parent string, projectIds, locations []string, assetTypesData []byte) ([]*assetpb.ResourceSearchResult, error) {
		return []*assetpb.ResourceSearchResult{
			{
				Name:      "//run.googleapis.com/projects/p1/locations/us-central1/services/svc1",
				AssetType: "run.googleapis.com/Service",
				Location:  "us-central1",
			},
		}, nil
	})
	defer cleanupSearchProj()

	containsVal := "svc"
	tests := []struct {
		name     string
		selector Selector
	}{
		{
			name:     "autoDetect selector",
			selector: Selector{AutoDetect: true},
		},
		{
			name:     "perK8sNamespace selector",
			selector: Selector{PerK8sNamespace: true},
		},
		{
			name:     "perK8sAppLabel selector",
			selector: Selector{PerK8sAppLabel: true},
		},
		{
			name: "projectKeys selector",
			selector: Selector{
				ProjectKeys: &ProjectKey{AppName: "projapp", ProjectIds: []string{"p1"}},
			},
		},
		{
			name: "label selector",
			selector: Selector{
				Label: &KeyValue{Key: "env", Value: "prod"},
			},
		},
		{
			name: "tag selector",
			selector: Selector{
				Tag: &KeyValue{Key: "cost-center", Value: "123"},
			},
		},
		{
			name: "contains selector",
			selector: Selector{
				Contains: &containsVal,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqPayload := GenerateRequest{
				Selector: tt.selector,
				Scope: Scope{
					Parent:            "projects/p1",
					ManagementProject: "p1",
					Locations:         []string{"us-central1"},
				},
				Action: Action{ReportOnly: true},
				Options: &Options{
					Attributes: &Attributes{
						Criticality: AttributesType{Type: "MISSION_CRITICAL"},
					},
				},
			}
			body, _ := json.Marshal(reqPayload)

			req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			generateHandler(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("generateHandler (%s) returned status %d, want %d: %s", tt.name, rr.Code, http.StatusOK, rr.Body.String())
			}
		})
	}
}

func TestGenerateHandlerFailureFromClient(t *testing.T) {
	cleanupClient := client.SetAppHubClientFuncForTest(func() (client.AppHubClient, error) {
		return &testAppHubClient{}, nil
	})
	defer cleanupClient()

	cleanupSearch := client.SetSearchAssetsFuncForTest(func(parent, labelKey, labelValue, tagKey, tagValue, contains string, locations []string, assetTypesData []byte) ([]*assetpb.ResourceSearchResult, error) {
		return nil, fmt.Errorf("search failed with internal error")
	})
	defer cleanupSearch()

	reqPayload := GenerateRequest{
		Selector: Selector{
			Label: &KeyValue{Key: "env", Value: "prod"},
		},
		Scope: Scope{
			Parent:            "projects/p1",
			ManagementProject: "p1",
			Locations:         []string{"us-central1"},
		},
		Action: Action{
			ReportOnly: true,
		},
	}
	body, _ := json.Marshal(reqPayload)

	req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()

	generateHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("generateHandler returned status %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}
