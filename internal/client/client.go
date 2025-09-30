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

	apphubpb "cloud.google.com/go/apphub/apiv1/apphubpb"
	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	"google.golang.org/api/iterator"
)

var (
	searchAssetsFunc    = searchAssets
	getAppHubClientFunc = getAppHubClient
)

func GenerateAppsAssetInventory(projectID, managementProject, labelKey, labelValue, tagKey, tagValue,
	contains string, locations []string, attributesData, assetTypesData []byte,
) error {
	logger := clilog.GetLogger()
	var appLocation string

	logger.Info("Running CAIS Search with location and Filters")
	assets, err := searchAssetsFunc(projectID, labelKey, labelValue, tagKey, tagValue, contains, locations, assetTypesData)
	if err != nil {
		return fmt.Errorf("error searching assets: %w", err)
	}

	if len(assets) == 0 {
		logger.Warn("No assets found that matched the filter")
		return fmt.Errorf("no assets found that matched the filter")
	}

	logger.Info("Found assets to process", "count", len(assets))

	apphubClient, err := getAppHubClientFunc()
	if err != nil {
		return fmt.Errorf("error getting apphub client: %w", err)
	}

	defer closeAppHubClient(apphubClient)

	if len(locations) > 1 {
		appLocation = "global"
	} else {
		appLocation = locations[0]
	}

	// For each asset returned
	for _, asset := range assets {
		logger.Info("Processing asset", "assetName", asset.Name, "assetType", asset.AssetType)

		var discoveredName, appName string

		// Identity if it is a service or workload
		appHubType := identifyServiceOrWorkload(asset.AssetType)

		// Lookup App Hub to get the discovered name
		if discoveredName, err = lookupDiscoveredServiceOrWorkload(apphubClient, managementProject,
			asset.Location,
			asset.Name,
			appHubType); err != nil {
			logger.Warn("Discovered Service/Workload not found, perhaps already registered", "assetName", asset.Name, "error", err)
		}
		// If the discovered name is not empty,
		if discoveredName != "" {
			appName = getAppName(labelKey, tagKey, contains, asset)
			// create the application if it does not exist
			if _, err = getOrCreateAppHubApplication(apphubClient, managementProject, appLocation, appName, attributesData); err != nil {
				logger.Error("Failed to create or get application", "application", appName, "error", err)
				return fmt.Errorf("error creating application: %w", err)
			}
			displayName := asset.Name[strings.LastIndex(asset.Name, "/")+1:]

			// Registry the service or workload
			if err = registerServiceWithApplication(apphubClient, managementProject,
				appLocation,
				appName,
				discoveredName,
				displayName,
				appHubType,
				attributesData); err != nil {
				logger.Error("Failed to register service with application", "application", appName, "service", displayName, "error", err)
				return fmt.Errorf("error registering service: %w", err)
			}
		}
	}
	logger.Info("Successfully finished processing all assets.")
	return nil
}

func GenerateAppsCloudLogging(projectID, managementProject, logLabelKey, logLabelValue string, locations []string, attributesData []byte) error {
	logger := clilog.GetLogger()
	var appLocation string

	logger.Info("Running Cloud Logging with location and Filters")

	assets, err := filterLogs(projectID, logLabelKey, logLabelValue, locations)
	if err != nil {
		return fmt.Errorf("error searching logs: %w", err)
	}

	if len(assets) == 0 {
		logger.Warn("No assets found that matched the filter")
		return fmt.Errorf("no assets found that matched the filter")
	}

	logger.Info("Found assets from logs to process", "count", len(assets))

	apphubClient, err := getAppHubClientFunc()
	if err != nil {
		return fmt.Errorf("error getting apphub client: %w", err)
	}

	defer closeAppHubClient(apphubClient)

	if len(locations) > 1 {
		appLocation = "global"
	} else {
		appLocation = locations[0]
	}

	// For each asset returned
	for assetURI, asset := range assets {
		logger.Info("Processing asset from logs", "assetURI", assetURI, "assetName", asset.Name)

		var discoveredName, appName string

		// Lookup App Hub to get the discovered name
		if discoveredName, err = lookupDiscoveredServiceOrWorkload(apphubClient, managementProject,
			asset.Location,
			assetURI,
			asset.AppHubType); err != nil {
			logger.Warn("Discovered Service/Workload not found, perhaps already registered", "assetURI", assetURI, "error", err)
		}

		// If the discovered name is not empty,
		if discoveredName != "" {
			appName = logLabelValue
			// create the application if it does not exist
			if _, err = getOrCreateAppHubApplication(apphubClient, managementProject, appLocation, appName, attributesData); err != nil {
				logger.Error("Failed to create or get application", "application", appName, "error", err)
				return fmt.Errorf("error creating application: %w", err)
			}
			displayName := asset.Name

			// Registry the service or workload
			if err = registerServiceWithApplication(apphubClient, managementProject,
				appLocation,
				appName,
				discoveredName,
				displayName,
				asset.AppHubType,
				attributesData); err != nil {
				logger.Error("Failed to register service with application", "application", appName, "service", displayName, "error", err)
				return fmt.Errorf("error registering service: %w", err)
			}
		}
	}
	logger.Info("Successfully finished processing all assets from logs.")
	return nil
}

func DeleteAllApps(managementProject string, locations []string) error {
	logger := clilog.GetLogger()
	ctx := context.Background()
	apphubClient, err := getAppHubClientFunc()
	if err != nil {
		return fmt.Errorf("error getting apphub client: %w", err)
	}

	defer closeAppHubClient(apphubClient)

	logger.Info("Attempting deletion of applications")
	for _, location := range locations {
		// Parent format: projects/{project}/locations/{location}/applications/{application_id}
		parent := fmt.Sprintf("projects/%s/locations/%s", managementProject, location)
		req := &apphubpb.ListApplicationsRequest{
			Parent: parent,
		}

		// Call the ListApplications API
		listApplications := apphubClient.ListApplications(ctx, req)
		for {
			app, err := listApplications.Next()
			if err != nil {
				if err == iterator.Done {
					break
				}
				return fmt.Errorf("failed to list applications: %w", err)
			}

			appName := app.Name[strings.LastIndex(app.Name, "/")+1:]
			logger.Info("Deleting application", "application", appName, "location", location)
			if err = deleteApp(apphubClient, managementProject, location, appName); err != nil {
				return fmt.Errorf("error deleting application %s: %w", appName, err)
			}
		}
	}
	logger.Info("Successfully finished deleting applications.")
	return nil
}

func getAppName(labelKey, tagKey, contains string, asset *assetpb.ResourceSearchResult) string {
	if labelKey != "" {
		return asset.GetLabels()[labelKey]
	} else if tagKey != "" {
		for _, tag := range asset.GetTags() {
			if *tag.TagKey == tagKey {
				return *tag.TagValue
			}
		}
		return "unknown"
	} else {
		return contains
	}
}
