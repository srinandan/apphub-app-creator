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

func GenerateAppsByLabel(projectID, managementProject, region, labelKey, tagKey, contains string, data []byte) error {
	logger := clilog.GetLogger()

	logger.Info("\n--- Running Search with Region and Label Key Filter ---")
	assets, err := searchAssets(projectID, region, labelKey, tagKey, contains)
	if err != nil {
		return fmt.Errorf("error searching assets: %w", err)
	}

	apphubClient, err := getAppHubClient()
	if err != nil {
		return fmt.Errorf("error getting apphub client: %w", err)
	}

	defer closeAppHubClient(apphubClient)

	// For each asset returned
	for _, asset := range assets {
		var discoveredName string
		// Identity if it is a service or workload
		appHubType := identifyServiceOrWorkload(asset.AssetType)

		// Lookup App Hub to get the discovered name
		if discoveredName, err = lookupDiscoveredServiceOrWOrkload(apphubClient, managementProject,
			region,
			asset.Name,
			appHubType); err != nil {
			logger.Warn("Discovered Service not found, perhaps already registered")
		}
		// If the discovered name is not empty,
		if discoveredName != "" {
			labelValue := asset.GetLabels()[labelKey]
			// create the application if it does not exist
			if _, err = getOrCreateAppHubApplication(apphubClient, managementProject, region, labelValue, data); err != nil {
				return fmt.Errorf("error creating application: %w", err)
			}
			displayName := asset.Name[strings.LastIndex(asset.Name, "/")+1:]

			// Registry the service or workload
			if err = registerServiceWithApplication(apphubClient, managementProject,
				region,
				labelValue,
				discoveredName,
				displayName,
				appHubType,
				data); err != nil {
				return fmt.Errorf("error registering service: %w", err)
			}
		}
	}
	return nil
}
