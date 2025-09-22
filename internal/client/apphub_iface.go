
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

	apphub "cloud.google.com/go/apphub/apiv1"
	apphubpb "cloud.google.com/go/apphub/apiv1/apphubpb"
	"github.com/googleapis/gax-go/v2"
)

// appHubClient is an interface that wraps the apphub.Client.

type appHubClient interface {
	LookupDiscoveredService(ctx context.Context, req *apphubpb.LookupDiscoveredServiceRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredServiceResponse, error)
	LookupDiscoveredWorkload(ctx context.Context, req *apphubpb.LookupDiscoveredWorkloadRequest, opts ...gax.CallOption) (*apphubpb.LookupDiscoveredWorkloadResponse, error)
	GetApplication(ctx context.Context, req *apphubpb.GetApplicationRequest, opts ...gax.CallOption) (*apphubpb.Application, error)
	CreateApplication(ctx context.Context, req *apphubpb.CreateApplicationRequest, opts ...gax.CallOption) (*apphub.CreateApplicationOperation, error)
	CreateService(ctx context.Context, req *apphubpb.CreateServiceRequest, opts ...gax.CallOption) (*apphub.CreateServiceOperation, error)
	CreateWorkload(ctx context.Context, req *apphubpb.CreateWorkloadRequest, opts ...gax.CallOption) (*apphub.CreateWorkloadOperation, error)
	Close() error
}
