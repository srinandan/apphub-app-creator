
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
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	apphubpb "cloud.google.com/go/apphub/apiv1/apphubpb"
)

// mockAppHubClient is a mock of the App Hub client.

type mockAppHubClient struct {
	lookupDiscoveredServiceFunc  func(ctx context.Context, req *apphubpb.LookupDiscoveredServiceRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredServiceResponse, error)
	lookupDiscoveredWorkloadFunc func(ctx context.Context, req *apphubpb.LookupDiscoveredWorkloadRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredWorkloadResponse, error)
	getApplicationFunc           func(ctx context.Context, req *apphubpb.GetApplicationRequest, opts ...gax.CallOption) (*apphubpb.Application, error)
	createApplicationFunc        func(ctx context.Context, req *apphubpb.CreateApplicationRequest, opts ...gax.CallOption) (*apphub.CreateApplicationOperation, error)
	createServiceFunc            func(ctx context.Context, req *apphubpb.CreateServiceRequest, opts ...gax.CallOption) (*apphub.CreateServiceOperation, error)
	createWorkloadFunc           func(ctx context.Context, req *apphubpb.CreateWorkloadRequest, opts ...gax.CallOption) (*apphub.CreateWorkloadOperation, error)
}

func (m *mockAppHubClient) LookupDiscoveredService(ctx context.Context, req *apphubpb.LookupDiscoveredServiceRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredServiceResponse, error) {
	return m.lookupDiscoveredServiceFunc(ctx, req, opts...)
}

func (m *mockAppHubClient) LookupDiscoveredWorkload(ctx context.Context, req *apphubpb.LookupDiscoveredWorkloadRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredWorkloadResponse, error) {
	return m.lookupDiscoveredWorkloadFunc(ctx, req, opts...)
}

func (m *mockAppHubClient) GetApplication(ctx context.Context, req *apphubpb.GetApplicationRequest, opts ...gax.CallOption) (*apphubpb.Application, error) {
	return m.getApplicationFunc(ctx, req, opts...)
}

func (m *mockAppHubClient) CreateApplication(ctx context.Context, req *apphubpb.CreateApplicationRequest, opts ...gax.CallOption) (*apphub.CreateApplicationOperation, error) {
	return m.createApplicationFunc(ctx, req, opts...)
}

func (m *mockAppHubClient) CreateService(ctx context.Context, req *apphubpb.CreateServiceRequest, opts ...gax.CallOption) (*apphub.CreateServiceOperation, error) {
	return m.createServiceFunc(ctx, req, opts...)
}

func (m *mockAppHubClient) CreateWorkload(ctx context.Context, req *apphubpb.CreateWorkloadRequest, opts ...gax.CallOption) (*apphub.CreateWorkloadOperation, error) {
	return m.createWorkloadFunc(ctx, req, opts...)
}

func (m *mockAppHubClient) Close() error {
	return nil
}

// mockCreateApplicationOperation is a mock of the CreateApplicationOperation.
type mockCreateApplicationOperation struct {
	apphub.CreateApplicationOperation
	waitFunc func(context.Context) (*apphubpb.Application, error)
}

func (m *mockCreateApplicationOperation) Wait(ctx context.Context, opts ...gax.CallOption) (*apphubpb.Application, error) {
	return m.waitFunc(ctx)
}

func (m *mockCreateApplicationOperation) Name() string {
	return "mock-operation"
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
			name:       "Permission Denied",
			appHubType: "discoveredService",
			mockClient: &mockAppHubClient{
				lookupDiscoveredServiceFunc: func(ctx context.Context, req *apphubpb.LookupDiscoveredServiceRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredServiceResponse, error) {
					return nil, status.Error(codes.PermissionDenied, "permission denied")
				},
			},
			wantErr:       true,
			expectedError: "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := lookupDiscoveredServiceOrWorkload(tt.mockClient, "test-project", "test-region", "test-uri", tt.appHubType)

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

/*
	tests := []struct {
		name       string
		mockClient appHubClient
		wantApp    *apphubpb.Application
		wantErr    bool
	}{
		{
			name: "Application Exists",
			mockClient: &mockAppHubClient{
				getApplicationFunc: func(ctx context.Context, req *apphubpb.GetApplicationRequest, opts ...gax.CallOption) (*apphubpb.Application, error) {
					return &apphubpb.Application{Name: "existing-app"}, nil
				},
			},
			wantApp: &apphubpb.Application{Name: "existing-app"},
			wantErr: false,
		},
		{
			name: "Application Created",
			mockClient: &mockAppHubClient{
				getApplicationFunc: func(ctx context.Context, req *apphubpb.GetApplicationRequest, opts ...gax.CallOption) (*apphubpb.Application, error) {
					return nil, status.Error(codes.NotFound, "not found")
				},
				createApplicationFunc: func(ctx context.Context, req *apphubpb.CreateApplicationRequest, opts ...gax.CallOption) (*apphub.CreateApplicationOperation, error) {
					return &apphub.CreateApplicationOperation{},
 nil
				},
			},
			wantApp: &apphubpb.Application{Name: "new-app"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := getOrCreateAppHubApplication(tt.mockClient, "test-project", "test-region", "test-app", nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("getOrCreateAppHubApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if app.Name != tt.wantApp.Name {
				t.Errorf("getOrCreateAppHubApplication() = %v, want %v", app.Name, tt.wantApp.Name)
			}
		})
	}
}

// mockLRO is a mock long-running operation.
type mockLRO struct {
	waitFunc func(context.Context) (interface{}, error)
}

func (m *mockLRO) Wait(ctx context.Context, opts ...gax.CallOption) (interface{}, error) {
	return m.waitFunc(ctx)
}

func (m *mockLRO) Name() string {
	return "mock-lro"
}

func (m *mockLRO) Metadata() (*longrunningpb.Operation, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockLRO) Done() bool {
	return false
}

func (m *mockLRO) Poll(ctx context.Context, opts ...gax.CallOption) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockLRO) Cancel(ctx context.Context, opts ...gax.CallOption) error {
	return fmt.Errorf("not implemented")
}

func (m *mockLRO) Delete(ctx context.Context, opts ...gax.CallOption) error {
	return fmt.Errorf("not implemented")
}

*/
