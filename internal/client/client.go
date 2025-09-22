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
)

var searchAssetsFunc = searchAssets
var getAppHubClientFunc = getAppHubClient

func GenerateApps(projectID, managementProject, labelKey, labelValue, tagKey, tagValue, contains string, locations []string, attributesData, assetTypesData []byte) error {
	logger := clilog.GetLogger()
	var appLocation string

	logger.Info("Running Search with location and Filters")
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
		var discoveredName string
		// Identity if it is a service or workload
		appHubType := identifyServiceOrWorkload(asset.AssetType)

		// Lookup App Hub to get the discovered name
		if discoveredName, err = lookupDiscoveredServiceOrWorkload(apphubClient, managementProject,
			asset.Location,
			asset.Name,
			appHubType); err != nil {
			logger.Warn("Discovered Service not found, perhaps already registered")
		}
		// If the discovered name is not empty,
		if discoveredName != "" {
			labelValue := asset.GetLabels()[labelKey]
			// create the application if it does not exist
			if _, err = getOrCreateAppHubApplication(apphubClient, managementProject, appLocation, labelValue, attributesData); err != nil {
				return fmt.Errorf("error creating application: %w", err)
			}
			displayName := asset.Name[strings.LastIndex(asset.Name, "/")+1:]

			// Registry the service or workload
			if err = registerServiceWithApplication(apphubClient, managementProject,
				appLocation,
				labelValue,
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
