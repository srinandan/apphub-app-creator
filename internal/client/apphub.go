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
	"internal/clilog"
	"strings"

	apphub "cloud.google.com/go/apphub/apiv1"
	apphubpb "cloud.google.com/go/apphub/apiv1/apphubpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// lookupDiscoveredService finds a DiscoveredService or Workload resource in App Hub based on its underlying resource URI.
// The DiscoveredService/Workload represents an existing GCP resource (like a Cloud Run service) that App Hub is aware of.
func lookupDiscoveredServiceOrWOrkload(apiclient *apphub.Client, projectID, region, resourceURI, appHubType string) (string, error) {
	ctx := context.Background()

	logger := clilog.GetLogger()

	// Construct the parent path and request
	// Parent format: projects/{project}/locations/{location}
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, region)

	if appHubType == "discoveredService" {

		req := &apphubpb.LookupDiscoveredServiceRequest{
			Parent: parent,
			Uri:    resourceURI,
		}

		logger.Info("Looking up Discovered Service for URI", parent, resourceURI)

		// Call the LookupDiscoveredService API
		response, err := apiclient.LookupDiscoveredService(ctx, req)
		if err != nil {
			// Check for specific gRPC error codes
			if st, ok := status.FromError(err); ok {
				if st.Code() == codes.PermissionDenied {
					return "", fmt.Errorf("permission denied: ensure the user has the 'apphub.discoveredServices.list' permission on the project: %w", err)
				}
				// Other errors could include InvalidArgument if the URI format is wrong
				return "", fmt.Errorf("app hub lookup API failed (Code: %s): %w", st.Code().String(), err)
			}
			return "", fmt.Errorf("app hub lookup API failed: %w", err)
		}

		// Check if a Discovered Service was returned and return its Name
		discoveredService := response.GetDiscoveredService()
		if discoveredService == nil {
			return "", fmt.Errorf("discovered service not found for URI: %s", resourceURI)
		}

		return discoveredService.GetName(), nil
	} else {
		req := &apphubpb.LookupDiscoveredWorkloadRequest{
			Parent: parent,
			Uri:    resourceURI,
		}

		logger.Info("Looking up Workload in '%s' for URI: '%s'\n", parent, resourceURI)

		// 3. Call the LookupWorkload API
		response, err := apiclient.LookupDiscoveredWorkload(ctx, req)
		if err != nil {
			// Check for specific gRPC error codes
			if st, ok := status.FromError(err); ok {
				if st.Code() == codes.PermissionDenied {
					return "", fmt.Errorf("permission denied: ensure the user has the 'apphub.workloads.list' permission on the project: %w", err)
				}
				return "", fmt.Errorf("app hub lookup API failed (Code: %s): %w", st.Code().String(), err)
			}
			return "", fmt.Errorf("app hub lookup API failed: %w", err)
		}

		// 4. Check if a Workload was returned and return its Name
		workload := response.GetDiscoveredWorkload()
		if workload == nil {
			return "", fmt.Errorf("workload not found for URI: %s", resourceURI)
		}

		return workload.GetName(), nil
	}
}

// getOrCreateAppHubApplication attempts to retrieve an App Hub application by name.
// If it does not exist, it creates a new one and waits for the operation to complete.
func getOrCreateAppHubApplication(apiclient *apphub.Client, projectID, region, appID string, data []byte) (*apphubpb.Application, error) {
	ctx := context.Background()

	logger := clilog.GetLogger()

	// Construct the full resource name for the Application
	// Name format: projects/{project}/locations/{location}/applications/{application_id}
	applicationName := fmt.Sprintf("projects/%s/locations/%s/applications/%s", projectID, region, appID)
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, region)

	// Check if the Application already exists (GET call) ---
	getApplicationReq := &apphubpb.GetApplicationRequest{
		Name: applicationName,
	}

	app, err := apiclient.GetApplication(ctx, getApplicationReq)
	if err == nil {
		logger.Info("Application already exists. Returning existing resource.", "app-name", applicationName)
		return app, nil
	}

	// If the error is NOT_FOUND, proceed to create it.
	if st, ok := status.FromError(err); !ok || st.Code() != codes.NotFound {
		return nil, fmt.Errorf("failed to check for existing application '%s': %w", applicationName, err)
	}

	attr, err := newAttributesFromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse attributes: %w", err)
	}

	logger.Info("Application not found. Creating new application...", "app-name", applicationName)

	// Create the Application (CREATE call, which returns an LRO) ---
	createApplicationReq := &apphubpb.CreateApplicationRequest{
		Parent:        parent,
		ApplicationId: appID,
		Application: &apphubpb.Application{
			DisplayName: appID,
			// Set mandatory scope and optional attributes
			Scope: &apphubpb.Scope{
				Type: apphubpb.Scope_REGIONAL,
			},
			Attributes: attr,
		},
	}

	op, err := apiclient.CreateApplication(ctx, createApplicationReq)
	if err != nil {
		return nil, fmt.Errorf("failed to start application creation: %w", err)
	}

	logger.Info("Application creation started (Operation: %s). Waiting for completion...", "op-name", op.Name())

	// Wait function from the LRO client. This blocks until the operation is Done.
	createdApp, err := op.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("application creation failed during wait: %w", err)
	}

	logger.Info("Application successfully created.", "app-name", createdApp.Name)
	return createdApp, nil
}

// registerServiceWithApplication registers a Discovered Service as an App Hub Service
// within a specified Application.
func registerServiceWithApplication(apiclient *apphub.Client, projectID, region, appID, discoveredName, displayName, appHubType string, data []byte) error {
	ctx := context.Background()

	logger := clilog.GetLogger()

	// Determine the Service Parent (The Application Path)
	// Parent format: projects/{project}/locations/{location}/applications/{application_id}
	parent := fmt.Sprintf("projects/%s/locations/%s/applications/%s", projectID, region, appID)

	// Determine the Service ID from the Discovered Service Name.
	// Discovered Service Name format: projects/{p}/locations/{r}/discoveredServices/{ds_id}
	// We use the ds_id as the Service ID.
	parts := strings.Split(discoveredName, "/")
	if len(parts) < 6 {
		return fmt.Errorf("invalid discovered name format: %s", discoveredName)
	}
	id := parts[5] // The ID is the 6th element in the path array (0-indexed)

	// Construct the CreateService Request
	logger.Info("Registering into Application", appHubType, id, "app-name", appID)

	attr, err := newAttributesFromBytes(data)
	if err != nil {
		return fmt.Errorf("failed to parse attributes: %w", err)
	}

	if appHubType == "discoveredService" {

		req := &apphubpb.CreateServiceRequest{
			Parent:    parent,
			ServiceId: id,
			Service: &apphubpb.Service{
				// This links the new App Hub Service resource to the existing Discovered Service
				DiscoveredService: discoveredName,
				DisplayName:       displayName,
				Attributes:        attr,
			},
		}

		// Call the CreateService API (LRO)
		op, err := apiclient.CreateService(ctx, req)
		if err != nil {
			// Check for ALREADY_EXISTS if the service is already registered to this app
			if st, ok := status.FromError(err); ok && st.Code() == codes.AlreadyExists {
				logger.Info("Service is already registered with application. Skipping creation", "service", id, "app-name", appID)
				return nil
			}
			return fmt.Errorf("failed to start service registration: %w", err)
		}

		logger.Info("Service registration started. Waiting for completion...", "op-name", op.Name())

		// Wait for the LRO to complete
		createdService, err := op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("service registration failed during wait: %w", err)
		}

		logger.Info("Service successfully registered to application.", "service", createdService.Name, "app-name", appID)
		return nil
	} else {
		req := &apphubpb.CreateWorkloadRequest{
			Parent:     parent,
			WorkloadId: id,
			Workload: &apphubpb.Workload{
				// This links the new App Hub Service resource to the existing Discovered Workload
				DiscoveredWorkload: discoveredName,
				DisplayName:        displayName,
				Attributes:         attr,
			},
		}

		// Call the CreateWorkload API (LRO)
		op, err := apiclient.CreateWorkload(ctx, req)
		if err != nil {
			// Check for ALREADY_EXISTS if the workload is already registered to this app
			if st, ok := status.FromError(err); ok && st.Code() == codes.AlreadyExists {
				logger.Info("Workload is already registered with application. Skipping creation", "workload", id, "app-name", appID)
				return nil
			}
			return fmt.Errorf("failed to start workload registration: %w", err)
		}

		logger.Info("Workload registration started. Waiting for completion...", "op-name", op.Name())

		// Wait for the LRO to complete
		createdWorkload, err := op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("workload registration failed during wait: %w", err)
		}

		logger.Info("Workload successfully registered to application.", "workload", createdWorkload.Name, "app-name", appID)
		return nil
	}
}

func getAppHubClient() (*apphub.Client, error) {
	ctx := context.Background()

	apiclient, err := apphub.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create App Hub client: %w", err)
	}
	return apiclient, nil
}

func closeAppHubClient(apiclient *apphub.Client) {
	apiclient.Close()
}
