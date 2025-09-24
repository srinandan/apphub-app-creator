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
	"fmt"
	"internal/clilog"
	"strings"

	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
)

var searchAssetsFunc = searchAssets
var getAppHubClientFunc = getAppHubClient

func GenerateAppsAssetInventory(projectID, managementProject, labelKey, labelValue, tagKey, tagValue,
	contains string, locations []string, attributesData, assetTypesData []byte) error {

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

	apphubClient, err := getAppHubClientFunc()
	if err != nil {
		return fmt.Errorf("error getting apphub client: %w", err)
	}

	defer closeAppHubClient(apphubClient)

	if len(locations) > 0 {
		appLocation = "global"
	} else {
		appLocation = locations[0]
	}

	// For each asset returned
	for _, asset := range assets {
		var discoveredName, appName string

		// Identity if it is a service or workload
		appHubType := identifyServiceOrWorkload(asset.AssetType)

		// Lookup App Hub to get the discovered name
		if discoveredName, err = lookupDiscoveredServiceOrWorkload(apphubClient, managementProject,
			asset.Location,
			asset.Name,
			appHubType); err != nil {
			logger.Warn("Discovered Service/Workload not found, perhaps already registered")
		}
		// If the discovered name is not empty,
		if discoveredName != "" {
			appName = getAppName(labelKey, tagKey, contains, asset)
			// create the application if it does not exist
			if _, err = getOrCreateAppHubApplication(apphubClient, managementProject, appLocation, appName, attributesData); err != nil {
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
				return fmt.Errorf("error registering service: %w", err)
			}
		}
	}
	return nil
}

func GenerateAppsCloudLogging(projectID, managementProject, logLabelKey, logLabelValue string, locations []string, attributesData []byte) error {
	logger := clilog.GetLogger()
	var appLocation string

	logger.Info("Running Cloud Lgging with location and Filters")

	assets, err := filterLogs(projectID, logLabelKey, logLabelValue, locations)
	if err != nil {
		return fmt.Errorf("error searching logs: %w", err)
	}

	if len(assets) == 0 {
		logger.Warn("No assets found that matched the filter")
		return fmt.Errorf("no assets found that matched the filter")
	}

	apphubClient, err := getAppHubClientFunc()
	if err != nil {
		return fmt.Errorf("error getting apphub client: %w", err)
	}

	defer closeAppHubClient(apphubClient)

	if len(locations) > 0 {
		appLocation = "global"
	} else {
		appLocation = locations[0]
	}

	// For each asset returned
	for assetURI, asset := range assets {
		var discoveredName, appName string

		// Lookup App Hub to get the discovered name
		if discoveredName, err = lookupDiscoveredServiceOrWorkload(apphubClient, managementProject,
			asset.Location,
			assetURI,
			asset.AppHubType); err != nil {
			logger.Warn("Discovered Service/Workload not found, perhaps already registered")
		}

		// If the discovered name is not empty,
		if discoveredName != "" {
			appName = logLabelValue
			// create the application if it does not exist
			if _, err = getOrCreateAppHubApplication(apphubClient, managementProject, appLocation, appName, attributesData); err != nil {
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
				return fmt.Errorf("error registering service: %w", err)
			}
		}
	}
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
