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
	"crypto/sha256"
	"encoding/hex"
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

func GenerateAppsAssetInventory(parent, managementProject, labelKey, labelValue, tagKey, tagValue,
	contains string, locations []string, attributesData, assetTypesData []byte, reportOnly bool,
) (map[string][]string, error) {

	logger := clilog.GetLogger()
	var appLocation string
	generatedApplications := make(map[string][]string)

	logger.Info("Running CAIS Search with location and Filters")
	assets, err := searchAssetsFunc(parent, labelKey, labelValue, tagKey, tagValue, contains, locations, assetTypesData)
	if err != nil {
		return generatedApplications, fmt.Errorf("error searching assets: %w", err)
	}

	if len(assets) == 0 {
		logger.Warn("No assets found that matched the filter")
		return generatedApplications, fmt.Errorf("no assets found that matched the filter")
	}

	logger.Info("Found assets to process", "count", len(assets))

	apphubClient, err := getAppHubClientFunc()
	if err != nil {
		return generatedApplications, fmt.Errorf("error getting apphub client: %w", err)
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
			appHubType,
			asset); err != nil {
			logger.Warn("Discovered Service/Workload not found, perhaps already registered", "assetName", asset.Name, "error", err)
		}
		// If the discovered name is not empty,
		if discoveredName != "" {
			appName = getAppName(labelKey, tagKey, contains, labelValue, tagValue, asset)
			// store in array to generate report
			generatedApplications[appName] = []string{
				discoveredName,
				appHubType,
				asset.Name,
			}

			// perform the action is reportOnly is false
			if !reportOnly {
				// create the application if it does not exist
				if _, err = getOrCreateAppHubApplication(apphubClient, managementProject, appLocation, appName, attributesData); err != nil {
					logger.Error("Failed to create or get application", "application", appName, "error", err)
					return generatedApplications, fmt.Errorf("error creating application: %w", err)
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
					return generatedApplications, fmt.Errorf("error registering service: %w", err)
				}
			}
		}
	}
	logger.Info("Successfully finished processing all assets.")
	return generatedApplications, nil
}

func GenerateAppsCloudLogging(projectID, managementProject, logLabelKey, logLabelValue string,
	locations []string, attributesData []byte, reportOnly bool,
) (map[string][]string, error) {

	logger := clilog.GetLogger()
	var appLocation string
	generatedApplications := make(map[string][]string)

	logger.Info("Running Cloud Logging with location and Filters")

	assets, err := filterLogs(projectID, logLabelKey, logLabelValue, locations)
	if err != nil {
		return generatedApplications, fmt.Errorf("error searching logs: %w", err)
	}

	if len(assets) == 0 {
		logger.Warn("No assets found that matched the filter")
		return generatedApplications, fmt.Errorf("no assets found that matched the filter")
	}

	logger.Info("Found assets from logs to process", "count", len(assets))

	apphubClient, err := getAppHubClientFunc()
	if err != nil {
		return generatedApplications, fmt.Errorf("error getting apphub client: %w", err)
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
			asset.AppHubType, nil); err != nil {
			logger.Warn("Discovered Service/Workload not found, perhaps already registered", "assetURI", assetURI, "error", err)
		}

		// If the discovered name is not empty,
		if discoveredName != "" {
			appName = logLabelValue

			// store in array to generate report
			generatedApplications[appName] = []string{
				discoveredName,
				asset.AppHubType,
				asset.Name,
			}

			// perform the action is reportOnly is false
			if !reportOnly {
				// create the application if it does not exist
				if _, err = getOrCreateAppHubApplication(apphubClient, managementProject, appLocation, appName, attributesData); err != nil {
					logger.Error("Failed to create or get application", "application", appName, "error", err)
					return generatedApplications, fmt.Errorf("error creating application: %w", err)
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
					return generatedApplications, fmt.Errorf("error registering service: %w", err)
				}
			}
		}
	}
	logger.Info("Successfully finished processing all assets from logs.")
	return generatedApplications, nil
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

func GenerateAppsPerNamespace(parent, managementProject string, locations []string,
	attributesData []byte, reportOnly bool) (map[string][]string, error) {

	logger := clilog.GetLogger()
	var appLocation string
	generatedApplications := make(map[string][]string)

	logger.Info("Running CAIS Search with location and Filters")
	assets, err := searchKubernetes(parent, locations)
	if err != nil {
		return generatedApplications, fmt.Errorf("error searching assets: %w", err)
	}

	if len(assets) == 0 {
		logger.Warn("No assets found that matched the filter")
		return generatedApplications, fmt.Errorf("no assets found that matched the filter")
	}

	logger.Info("Found assets to process", "count", len(assets))

	apphubClient, err := getAppHubClientFunc()
	if err != nil {
		return generatedApplications, fmt.Errorf("error getting apphub client: %w", err)
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
			appHubType,
			asset); err != nil {
			logger.Warn("Discovered Service/Workload not found, perhaps already registered", "assetName", asset.Name, "error", err)
		}
		// If the discovered name is not empty,
		if discoveredName != "" {
			appName = getAppNameForKubernetes(asset.ParentFullResourceName)

			// store in array to generate report
			generatedApplications[appName] = []string{
				discoveredName,
				appHubType,
				asset.Name,
			}

			// perform the action is reportOnly is false
			if !reportOnly {
				// create the application if it does not exist
				if _, err = getOrCreateAppHubApplication(apphubClient, managementProject, appLocation, appName, attributesData); err != nil {
					logger.Error("Failed to create or get application", "application", appName, "error", err)
					return generatedApplications, fmt.Errorf("error creating application: %w", err)
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
					return generatedApplications, fmt.Errorf("error registering service: %w", err)
				}
			}
		}
	}
	logger.Info("Successfully finished processing all assets.")
	return generatedApplications, nil
}

func GenerateKubernetesApps(parent, managementProject string, locations []string, attributesData []byte,
	reportOnly bool) (map[string][]string, error) {
	logger := clilog.GetLogger()
	var appLocation string
	generatedApplications := make(map[string][]string)

	logger.Info("Running CAIS Search with location and Filters")
	assets, err := searchKubernetesApps(parent, locations)
	if err != nil {
		return generatedApplications, fmt.Errorf("error searching assets: %w", err)
	}

	if len(assets) == 0 {
		logger.Warn("No assets found that matched the filter")
		return generatedApplications, fmt.Errorf("no assets found that matched the filter")
	}

	logger.Info("Found assets to process", "count", len(assets))

	apphubClient, err := getAppHubClientFunc()
	if err != nil {
		return generatedApplications, fmt.Errorf("error getting apphub client: %w", err)
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
			appHubType,
			asset); err != nil {
			logger.Warn("Discovered Service/Workload not found, perhaps already registered", "assetName", asset.Name, "error", err)
		}
		// If the discovered name is not empty,
		if discoveredName != "" {
			appName = asset.GetLabels()[K8S_APP_LABEL]

			// store in array to generate report
			generatedApplications[appName] = []string{
				discoveredName,
				appHubType,
				asset.Name,
			}

			// perform the action is reportOnly is false
			if !reportOnly {
				// create the application if it does not exist
				if _, err = getOrCreateAppHubApplication(apphubClient, managementProject, appLocation, appName, attributesData); err != nil {
					logger.Error("Failed to create or get application", "application", appName, "error", err)
					return generatedApplications, fmt.Errorf("error creating application: %w", err)
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
					return generatedApplications, fmt.Errorf("error registering service: %w", err)
				}
			}
		}
	}
	logger.Info("Successfully finished processing all assets.")
	return generatedApplications, nil
}

func DeleteApp(managementProject, name string, locations []string) error {
	logger := clilog.GetLogger()
	apphubClient, err := getAppHubClientFunc()
	if err != nil {
		return fmt.Errorf("error getting apphub client: %w", err)
	}

	defer closeAppHubClient(apphubClient)

	logger.Info("Attempting deletion of application " + name)
	for _, location := range locations {
		if err = deleteApp(apphubClient, managementProject, location, name); err != nil {
			return fmt.Errorf("error deleting application %s: %w", name, err)
		}
	}
	logger.Info("Successfully finished deleting application.")
	return nil
}

func getAppName(labelKey, tagKey, contains, labelValue, tagValue string, asset *assetpb.ResourceSearchResult) string {
	logger := clilog.GetLogger()
	if labelValue != "" {
		return labelValue
	} else if tagValue != "" {
		return tagValue
	} else if labelKey != "" {
		return asset.GetLabels()[labelKey]
	} else if tagKey != "" {
		for _, tag := range asset.GetTags() {
			lastElement := tag.GetTagKey()[strings.LastIndex(tag.GetTagKey(), "/")+1:]
			if lastElement == tagKey {
				return tag.GetTagValue()[strings.LastIndex(tag.GetTagValue(), "/")+1:]
			}
		}
		for _, effectiveTagDetails := range asset.GetEffectiveTags() {
			for _, tag := range effectiveTagDetails.GetEffectiveTags() {
				lastElement := tag.GetTagKey()[strings.LastIndex(tag.GetTagKey(), "/")+1:]
				if lastElement == tagKey {
					return tag.GetTagValue()[strings.LastIndex(tag.GetTagValue(), "/")+1:]
				}
			}
		}
		logger.Warn("unable to derive an application name, using unknown")
		return "unknown"
	} else {
		return contains
	}
}

func getAppNameForKubernetes(s string) string {
	namespace := s[strings.LastIndex(s, "/")+1:]
	project := strings.Split(s, "/")[2]
	cluster := strings.Split(s, "/")[5]
	return fmt.Sprintf("%s-%s-%s", namespace, createShortSHA(cluster), createShortSHA(project))
}

// createShortSHA generates a truncated SHA-256 hash of a string.
func createShortSHA(input string) string {
	// Create a new SHA-256 hasher.
	hasher := sha256.New()

	// Write the input string to the hasher.
	hasher.Write([]byte(input))

	// Get the full hash sum as a byte slice.
	fullHash := hasher.Sum(nil)

	// Convert the byte slice to a hexadecimal string.
	fullHashStr := hex.EncodeToString(fullHash)

	// Truncate the string to the desired length (e.g., 7 characters).
	shortHash := fullHashStr[:7]

	return shortHash
}
