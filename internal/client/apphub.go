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
	"regexp"
	"strings"

	apphub "cloud.google.com/go/apphub/apiv1"
	apphubpb "cloud.google.com/go/apphub/apiv1/apphubpb"
	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// lookupDiscoveredService finds a DiscoveredService or Workload resource in App Hub based on its underlying resource URI.
// The DiscoveredService/Workload represents an existing GCP resource (like a Cloud Run service) that App Hub is aware of.
func lookupDiscoveredServiceOrWorkload(apiclient appHubClient, projectID, location, resourceURI, appHubType string, asset *assetpb.ResourceSearchResult) (string, error) {
	ctx := context.Background()
	logger := clilog.GetLogger()

	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)

	var (
		name string
		err  error
	)

	switch appHubType {
	case "discoveredService":
		fixedResourceURI := fixResourceURI(resourceURI, asset)
		req := &apphubpb.LookupDiscoveredServiceRequest{
			Parent: parent,
			Uri:    fixedResourceURI,
		}
		logger.Info("Looking up Discovered Service for URI", "parent", parent, "uri", fixedResourceURI)
		var response *apphubpb.LookupDiscoveredServiceResponse
		response, err = apiclient.LookupDiscoveredService(ctx, req)
		if err == nil {
			if response.GetDiscoveredService() == nil {
				logger.Warn("Lookup API succeeded but returned no discovered service", "uri", fixedResourceURI)
				return "", fmt.Errorf("discovered service not found for URI: %s", fixedResourceURI)
			}
			name = response.GetDiscoveredService().GetName()
		}

	case "discoveredWorkload":
		req := &apphubpb.LookupDiscoveredWorkloadRequest{
			Parent: parent,
			Uri:    resourceURI,
		}
		logger.Info("Looking up Workload in", "parent", parent, "uri", resourceURI)
		var response *apphubpb.LookupDiscoveredWorkloadResponse
		response, err = apiclient.LookupDiscoveredWorkload(ctx, req)
		if err == nil {
			if response.GetDiscoveredWorkload() == nil {
				logger.Warn("Lookup API succeeded but returned no discovered workload", "uri", resourceURI)
				return "", fmt.Errorf("workload not found for URI: %s", resourceURI)
			}
			name = response.GetDiscoveredWorkload().GetName()
		}
	default:
		return "", fmt.Errorf("invalid appHubType: %s", appHubType)
	}

	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.PermissionDenied {
				permission := "apphub.discoveredServices.list"
				if appHubType == "discoveredWorkload" {
					permission = "apphub.workloads.list"
				}
				return "", fmt.Errorf("permission denied: ensure the user has the '%s' permission on the project: %w", permission, err)
			}
			logger.Error("App Hub lookup API failed", "code", st.Code().String(), "error", err)
			return "", fmt.Errorf("app hub lookup API failed (Code: %s): %w", st.Code().String(), err)
		}
		return "", fmt.Errorf("app hub lookup API failed: %w", err)
	}

	logger.Info("Successfully found discovered resource", "name", name, "type", appHubType)
	return name, nil
}

// getOrCreateAppHubApplication attempts to retrieve an App Hub application by name.
// If it does not exist, it creates a new one and waits for the operation to complete.
func getOrCreateAppHubApplication(apiclient appHubClient, projectID, location, appID string, data []byte) (*apphubpb.Application, error) {
	ctx := context.Background()

	logger := clilog.GetLogger()

	var appScope apphubpb.Scope_Type

	// Construct the full resource name for the Application
	// Name format: projects/{project}/locations/{location}/applications/{application_id}
	applicationName := fmt.Sprintf("projects/%s/locations/%s/applications/%s", projectID, location, appID)
	parent := fmt.Sprintf("projects/%s/locations/%s", projectID, location)

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
		logger.Error("Failed to check for existing application", "app-name", applicationName, "error", err)
		return nil, fmt.Errorf("failed to check for existing application '%s': %w", applicationName, err)
	}

	attr, err := newAttributesFromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse attributes: %w", err)
	}

	logger.Info("Application not found. Creating new application...", "app-name", applicationName)

	if location == "global" {
		appScope = apphubpb.Scope_GLOBAL
	} else {
		appScope = apphubpb.Scope_REGIONAL
	}

	// Create the Application (CREATE call, which returns an LRO) ---
	createApplicationReq := &apphubpb.CreateApplicationRequest{
		Parent:        parent,
		ApplicationId: appID,
		Application: &apphubpb.Application{
			DisplayName: appID,
			// Set mandatory scope and optional attributes
			Scope: &apphubpb.Scope{
				Type: appScope,
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
func registerServiceWithApplication(apiclient appHubClient, projectID, location, appID, discoveredName, displayName, appHubType string, data []byte) error {
	ctx := context.Background()

	logger := clilog.GetLogger()

	// Determine the Service Parent (The Application Path)
	// Parent format: projects/{project}/locations/{location}/applications/{application_id}
	parent := fmt.Sprintf("projects/%s/locations/%s/applications/%s", projectID, location, appID)

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
			// Check for ALREADY_EXISTS if the workload is already registered to this app
			if st, ok := status.FromError(err); ok && st.Code() == codes.FailedPrecondition {
				logger.Info("Service is already registered with application. Skipping creation", "service", id, "app-name", appID)
				return nil
			}
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
			// Check for ALREADY_EXISTS if the workload is already registered to this app
			if st, ok := status.FromError(err); ok && st.Code() == codes.FailedPrecondition {
				logger.Info("Workload is already registered with application. Skipping creation", "workload", id, "app-name", appID)
				return nil
			}
			return fmt.Errorf("workload registration failed during wait: %w", err)
		}

		logger.Info("Workload successfully registered to application.", "workload", createdWorkload.Name, "app-name", appID)
		return nil
	}
}

func removeAllServices(apiclient appHubClient, projectID, location, appID string) error {
	ctx := context.Background()

	logger := clilog.GetLogger()

	// Determine the Service Parent (The Application Path)
	// Parent format: projects/{project}/locations/{location}/applications/{application_id}
	parent := fmt.Sprintf("projects/%s/locations/%s/applications/%s", projectID, location, appID)

	reqServices := &apphubpb.ListServicesRequest{
		Parent: parent,
	}

	// Call the ListServices API
	listServices := apiclient.ListServices(ctx, reqServices)

	for {
		service, err := listServices.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return fmt.Errorf("failed to list services: %w", err)
		}

		// Construct the DeleteService Request
		reqDeleteService := &apphubpb.DeleteServiceRequest{
			Name: service.GetName(),
		}
		// Call the DeleteService API (LRO)
		op, err := apiclient.DeleteService(ctx, reqDeleteService)
		if err != nil {
			return fmt.Errorf("failed to start service deletion: %w", err)
		}

		op.Wait(ctx)
		logger.Info("Service successfully deleted.", "service", service.Name)
	}

	return nil
}

func removeAllWorkloads(apiclient appHubClient, projectID, location, appID string) error {
	ctx := context.Background()

	logger := clilog.GetLogger()

	// Determine the Service Parent (The Application Path)
	// Parent format: projects/{project}/locations/{location}/applications/{application_id}
	parent := fmt.Sprintf("projects/%s/locations/%s/applications/%s", projectID, location, appID)

	reqWorkloads := &apphubpb.ListWorkloadsRequest{
		Parent: parent,
	}

	// Call the ListWorkloads API
	listWorkloads := apiclient.ListWorkloads(ctx, reqWorkloads)

	for {
		workload, err := listWorkloads.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return fmt.Errorf("failed to list workloads: %w", err)
		}

		// Construct the DeleteWorkload Request
		reqDeleteWorkload := &apphubpb.DeleteWorkloadRequest{
			Name: workload.GetName(),
		}
		// Call the DeleteWorkload API (LRO)
		op, err := apiclient.DeleteWorkload(ctx, reqDeleteWorkload)
		if err != nil {
			return fmt.Errorf("failed to start workload deletion: %w", err)
		}

		op.Wait(ctx)
		logger.Info("Successfully deleted workload", "workload", workload.Name)
	}

	return nil
}

func deleteApp(apiclient appHubClient, projectID, location, appID string) error {
	var err error

	ctx := context.Background()

	logger := clilog.GetLogger()

	logger.Info("Removing all services from application", "app-name", appID)
	err = removeAllServices(apiclient, projectID, location, appID)
	if err != nil {
		return fmt.Errorf("failed to remove all services: %w", err)
	}

	logger.Info("Removing all workloads from application", "app-name", appID)
	err = removeAllWorkloads(apiclient, projectID, location, appID)
	if err != nil {
		return fmt.Errorf("failed to remove all workloads: %w", err)
	}

	// Parent format: projects/{project}/locations/{location}/applications/{application_id}
	parent := fmt.Sprintf("projects/%s/locations/%s/applications/%s", projectID, location, appID)

	req := &apphubpb.DeleteApplicationRequest{
		Name: parent,
	}

	// Delete the application
	op, err := apiclient.DeleteApplication(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to start application deletion: %w", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("application deletion failed during wait: %w", err)
	}
	logger.Info("Application successfully deleted", "app-name", appID)

	return nil
}

func getAppHubClient() (appHubClient, error) {
	ctx := context.Background()

	apiclient, err := apphub.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create App Hub client: %w", err)
	}
	return apiclient, nil
}

func closeAppHubClient(apiclient appHubClient) {
	apiclient.Close()
}

func fixResourceURI(resourceURI string, asset *assetpb.ResourceSearchResult) string {
	if asset == nil {
		return resourceURI
	}
	if asset.AssetType == "sqladmin.googleapis.com/Instance" {
		// step 1 replace the URI
		resourceURI = strings.Replace(resourceURI, "cloudsql.googleapis.com", "sqladmin.googleapis.com", 1)

		// step 2 get the project number
		projectNumber := strings.Split(asset.Project, "/")[1]

		// step 3 replace project id with project number
		re := regexp.MustCompile(`(projects/)([^/]+)(/instances/)`)
		resourceURI = re.ReplaceAllString(resourceURI, fmt.Sprintf("${1}%s${3}", projectNumber))
	}
	return resourceURI
}
