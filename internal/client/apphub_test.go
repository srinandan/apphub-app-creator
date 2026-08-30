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
	"strings"
	"testing"

	apphub "cloud.google.com/go/apphub/apiv1"
	apphubpb "cloud.google.com/go/apphub/apiv1/apphubpb"
	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockAppHubClient is a mock of the App Hub client.
type mockAppHubClient struct {
	lookupDiscoveredServiceFunc  func(ctx context.Context, req *apphubpb.LookupDiscoveredServiceRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredServiceResponse, error)
	lookupDiscoveredWorkloadFunc func(ctx context.Context, req *apphubpb.LookupDiscoveredWorkloadRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredWorkloadResponse, error)
	getApplicationFunc           func(ctx context.Context, req *apphubpb.GetApplicationRequest, opts ...gax.CallOption) (*apphubpb.Application, error)
	createApplicationFunc        func(ctx context.Context, req *apphubpb.CreateApplicationRequest, opts ...gax.CallOption) (*apphub.CreateApplicationOperation, error)
	createServiceFunc            func(ctx context.Context, req *apphubpb.CreateServiceRequest, opts ...gax.CallOption) (*apphub.CreateServiceOperation, error)
	createWorkloadFunc           func(ctx context.Context, req *apphubpb.CreateWorkloadRequest, opts ...gax.CallOption) (*apphub.CreateWorkloadOperation, error)
	deleteApplicationFunc        func(ctx context.Context, req *apphubpb.DeleteApplicationRequest, opts ...gax.CallOption) (*apphub.DeleteApplicationOperation, error)
	deleteServiceFunc            func(ctx context.Context, req *apphubpb.DeleteServiceRequest, opts ...gax.CallOption) (*apphub.DeleteServiceOperation, error)
	deleteWorkloadFunc           func(ctx context.Context, req *apphubpb.DeleteWorkloadRequest, opts ...gax.CallOption) (*apphub.DeleteWorkloadOperation, error)
	listDiscoveredServicesFunc   func(ctx context.Context, req *apphubpb.ListDiscoveredServicesRequest, opts ...gax.CallOption) *apphub.DiscoveredServiceIterator
	listDiscoveredWorkloadsFunc  func(ctx context.Context, req *apphubpb.ListDiscoveredWorkloadsRequest, opts ...gax.CallOption) *apphub.DiscoveredWorkloadIterator
}

func (m *mockAppHubClient) LookupDiscoveredService(ctx context.Context, req *apphubpb.LookupDiscoveredServiceRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredServiceResponse, error) {
	if m.lookupDiscoveredServiceFunc != nil {
		return m.lookupDiscoveredServiceFunc(ctx, req, opts...)
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (m *mockAppHubClient) LookupDiscoveredWorkload(ctx context.Context, req *apphubpb.LookupDiscoveredWorkloadRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredWorkloadResponse, error) {
	if m.lookupDiscoveredWorkloadFunc != nil {
		return m.lookupDiscoveredWorkloadFunc(ctx, req, opts...)
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (m *mockAppHubClient) GetApplication(ctx context.Context, req *apphubpb.GetApplicationRequest, opts ...gax.CallOption) (*apphubpb.Application, error) {
	if m.getApplicationFunc != nil {
		return m.getApplicationFunc(ctx, req, opts...)
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (m *mockAppHubClient) CreateApplication(ctx context.Context, req *apphubpb.CreateApplicationRequest, opts ...gax.CallOption) (*apphub.CreateApplicationOperation, error) {
	if m.createApplicationFunc != nil {
		return m.createApplicationFunc(ctx, req, opts...)
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (m *mockAppHubClient) CreateService(ctx context.Context, req *apphubpb.CreateServiceRequest, opts ...gax.CallOption) (*apphub.CreateServiceOperation, error) {
	if m.createServiceFunc != nil {
		return m.createServiceFunc(ctx, req, opts...)
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (m *mockAppHubClient) CreateWorkload(ctx context.Context, req *apphubpb.CreateWorkloadRequest, opts ...gax.CallOption) (*apphub.CreateWorkloadOperation, error) {
	if m.createWorkloadFunc != nil {
		return m.createWorkloadFunc(ctx, req, opts...)
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (m *mockAppHubClient) ListServices(ctx context.Context, req *apphubpb.ListServicesRequest, opts ...gax.CallOption) *apphub.ServiceIterator {
	return nil
}

func (m *mockAppHubClient) DeleteService(ctx context.Context, req *apphubpb.DeleteServiceRequest, opts ...gax.CallOption) (*apphub.DeleteServiceOperation, error) {
	if m.deleteServiceFunc != nil {
		return m.deleteServiceFunc(ctx, req, opts...)
	}
	return nil, nil
}

func (m *mockAppHubClient) ListWorkloads(ctx context.Context, req *apphubpb.ListWorkloadsRequest, opts ...gax.CallOption) *apphub.WorkloadIterator {
	return nil
}

func (m *mockAppHubClient) DeleteWorkload(ctx context.Context, req *apphubpb.DeleteWorkloadRequest, opts ...gax.CallOption) (*apphub.DeleteWorkloadOperation, error) {
	if m.deleteWorkloadFunc != nil {
		return m.deleteWorkloadFunc(ctx, req, opts...)
	}
	return nil, nil
}

func (m *mockAppHubClient) DeleteApplication(ctx context.Context, req *apphubpb.DeleteApplicationRequest, opts ...gax.CallOption) (*apphub.DeleteApplicationOperation, error) {
	if m.deleteApplicationFunc != nil {
		return m.deleteApplicationFunc(ctx, req, opts...)
	}
	return nil, nil
}

func (m *mockAppHubClient) ListApplications(ctx context.Context, req *apphubpb.ListApplicationsRequest, opts ...gax.CallOption) *apphub.ApplicationIterator {
	return nil
}

func (m *mockAppHubClient) ListDiscoveredServices(ctx context.Context, req *apphubpb.ListDiscoveredServicesRequest, opts ...gax.CallOption) *apphub.DiscoveredServiceIterator {
	if m.listDiscoveredServicesFunc != nil {
		return m.listDiscoveredServicesFunc(ctx, req, opts...)
	}
	return nil
}

func (m *mockAppHubClient) ListDiscoveredWorkloads(ctx context.Context, req *apphubpb.ListDiscoveredWorkloadsRequest, opts ...gax.CallOption) *apphub.DiscoveredWorkloadIterator {
	if m.listDiscoveredWorkloadsFunc != nil {
		return m.listDiscoveredWorkloadsFunc(ctx, req, opts...)
	}
	return nil
}

func (m *mockAppHubClient) Close() error {
	return nil
}

func TestLookupDiscoveredServiceOrWorkload(t *testing.T) {
	tests := []struct {
		name          string
		appHubType    string
		mockClient    appHubClient
		wantName      string
		wantErr       bool
		expectedError string
	}{
		{
			name:       "Lookup Discovered Service - Success",
			appHubType: "discoveredService",
			mockClient: &mockAppHubClient{
				lookupDiscoveredServiceFunc: func(ctx context.Context, req *apphubpb.LookupDiscoveredServiceRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredServiceResponse, error) {
					return &apphubpb.LookupDiscoveredServiceResponse{
						DiscoveredService: &apphubpb.DiscoveredService{
							Name: "test-service",
						},
					}, nil
				},
			},
			wantName: "test-service",
			wantErr:  false,
		},
		{
			name:       "Lookup Discovered Workload - Success",
			appHubType: "discoveredWorkload",
			mockClient: &mockAppHubClient{
				lookupDiscoveredWorkloadFunc: func(ctx context.Context, req *apphubpb.LookupDiscoveredWorkloadRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredWorkloadResponse, error) {
					return &apphubpb.LookupDiscoveredWorkloadResponse{
						DiscoveredWorkload: &apphubpb.DiscoveredWorkload{
							Name: "test-workload",
						},
					}, nil
				},
			},
			wantName: "test-workload",
			wantErr:  false,
		},
		{
			name:       "Permission Denied Service",
			appHubType: "discoveredService",
			mockClient: &mockAppHubClient{
				lookupDiscoveredServiceFunc: func(ctx context.Context, req *apphubpb.LookupDiscoveredServiceRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredServiceResponse, error) {
					return nil, status.Error(codes.PermissionDenied, "permission denied")
				},
			},
			wantErr:       true,
			expectedError: "permission denied",
		},
		{
			name:       "Permission Denied Workload",
			appHubType: "discoveredWorkload",
			mockClient: &mockAppHubClient{
				lookupDiscoveredWorkloadFunc: func(ctx context.Context, req *apphubpb.LookupDiscoveredWorkloadRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredWorkloadResponse, error) {
					return nil, status.Error(codes.PermissionDenied, "permission denied")
				},
			},
			wantErr:       true,
			expectedError: "permission denied",
		},
		{
			name:       "Invalid appHubType",
			appHubType: "invalidType",
			mockClient: &mockAppHubClient{},
			wantErr:    true,
		},
		{
			name:       "Service not found in response",
			appHubType: "discoveredService",
			mockClient: &mockAppHubClient{
				lookupDiscoveredServiceFunc: func(ctx context.Context, req *apphubpb.LookupDiscoveredServiceRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredServiceResponse, error) {
					return &apphubpb.LookupDiscoveredServiceResponse{}, nil
				},
			},
			wantErr: true,
		},
		{
			name:       "Workload not found in response",
			appHubType: "discoveredWorkload",
			mockClient: &mockAppHubClient{
				lookupDiscoveredWorkloadFunc: func(ctx context.Context, req *apphubpb.LookupDiscoveredWorkloadRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredWorkloadResponse, error) {
					return &apphubpb.LookupDiscoveredWorkloadResponse{}, nil
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := lookupDiscoveredServiceOrWorkload(tt.mockClient, "test-project", "test-region", "test-uri", tt.appHubType, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("lookupDiscoveredServiceOrWorkload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.expectedError != "" && !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("lookupDiscoveredServiceOrWorkload() error = %v, expectedError %v", err, tt.expectedError)
			}

			if name != tt.wantName {
				t.Errorf("lookupDiscoveredServiceOrWorkload() = %v, want %v", name, tt.wantName)
			}
		})
	}
}

func TestGetOrCreateAppHubApplication(t *testing.T) {
	tests := []struct {
		name       string
		mockClient appHubClient
		data       []byte
		wantErr    bool
	}{
		{
			name: "Application already exists",
			mockClient: &mockAppHubClient{
				getApplicationFunc: func(ctx context.Context, req *apphubpb.GetApplicationRequest, opts ...gax.CallOption) (*apphubpb.Application, error) {
					return &apphubpb.Application{Name: req.Name}, nil
				},
			},
			wantErr: false,
		},
		{
			name: "GetApplication returns unhandled error",
			mockClient: &mockAppHubClient{
				getApplicationFunc: func(ctx context.Context, req *apphubpb.GetApplicationRequest, opts ...gax.CallOption) (*apphubpb.Application, error) {
					return nil, status.Error(codes.PermissionDenied, "permission denied")
				},
			},
			wantErr: true,
		},
		{
			name: "Application not found but invalid attributes data",
			mockClient: &mockAppHubClient{
				getApplicationFunc: func(ctx context.Context, req *apphubpb.GetApplicationRequest, opts ...gax.CallOption) (*apphubpb.Application, error) {
					return nil, status.Error(codes.NotFound, "not found")
				},
			},
			data:    []byte("invalid json"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := getOrCreateAppHubApplication(tt.mockClient, "proj", "us-central1", "app1", tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("getOrCreateAppHubApplication() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && app == nil {
				t.Errorf("expected non-nil app")
			}
		})
	}
}

func TestRegisterServiceWithApplication(t *testing.T) {
	tests := []struct {
		name           string
		discoveredName string
		wantErr        bool
	}{
		{
			name:           "invalid discovered name format",
			discoveredName: "invalid-name",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registerServiceWithApplication(&mockAppHubClient{}, "proj", "us-central1", "app1", tt.discoveredName, "disp", "discoveredService", nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("registerServiceWithApplication() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFixResourceURI(t *testing.T) {
	tests := []struct {
		name        string
		resourceURI string
		asset       *assetpb.ResourceSearchResult
		want        string
	}{
		{
			name:        "nil asset returns URI unchanged",
			resourceURI: "//run.googleapis.com/projects/p1/locations/us-central1/services/svc1",
			asset:       nil,
			want:        "//run.googleapis.com/projects/p1/locations/us-central1/services/svc1",
		},
		{
			name:        "sqladmin asset replaces cloudsql and project ID with project number",
			resourceURI: "//cloudsql.googleapis.com/projects/my-project-id/instances/my-db",
			asset: &assetpb.ResourceSearchResult{
				AssetType: "sqladmin.googleapis.com/Instance",
				Project:   "projects/123456789",
			},
			want: "//sqladmin.googleapis.com/projects/123456789/instances/my-db",
		},
		{
			name:        "other asset type returns URI unchanged",
			resourceURI: "//run.googleapis.com/projects/p1/locations/us-central1/services/svc1",
			asset: &assetpb.ResourceSearchResult{
				AssetType: "run.googleapis.com/Service",
				Project:   "projects/123456789",
			},
			want: "//run.googleapis.com/projects/p1/locations/us-central1/services/svc1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fixResourceURI(tt.resourceURI, tt.asset)
			if got != tt.want {
				t.Errorf("fixResourceURI() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncateName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "short name",
			input: "my-service",
			want:  "my-service",
		},
		{
			name:  "exactly 63 characters",
			input: strings.Repeat("a", 63),
			want:  strings.Repeat("a", 63),
		},
		{
			name:  "longer than 63 characters gets truncated",
			input: strings.Repeat("a", 70),
			want:  strings.Repeat("a", 63),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateName(tt.input)
			if got != tt.want {
				t.Errorf("truncateName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetServiceWorkloadId(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		assetName string
		want      string
	}{
		{
			name:      "id with hyphen and normal asset name",
			id:        "prefix-12345",
			assetName: "my_service",
			want:      "my-service-12345",
		},
		{
			name:      "id without hyphen returns id directly",
			id:        "plainid",
			assetName: "my_service",
			want:      "plainid",
		},
		{
			name:      "long asset name truncated to 50 characters",
			id:        "prefix-suffix",
			assetName: strings.Repeat("a", 60),
			want:      strings.Repeat("a", 50) + "-suffix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getServiceWorkloadId(tt.id, tt.assetName)
			if got != tt.want {
				t.Errorf("getServiceWorkloadId() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCloseAppHubClient(t *testing.T) {
	// Should not panic on nil client or mock client
	closeAppHubClient(nil)
	closeAppHubClient(&mockAppHubClient{})
}
