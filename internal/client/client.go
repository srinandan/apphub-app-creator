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

func GenerateAppsByLabel(projectID, managementProject, region, labelKey string, data []byte) error {
	logger := clilog.GetLogger()

	logger.Info("\n--- Running Search with Region and Label Key Filter ---")
	assets, err := searchAssets(projectID, region, labelKey)
	if err != nil {
		return fmt.Errorf("error searching assets: %w", err)
	}

	for _, asset := range assets {
		var discoveredName string
		appHubType := identifyServiceOrWorkload(asset.AssetType)
		if discoveredName, err = lookupDiscoveredServiceOrWOrkload(managementProject,
			region,
			asset.Name,
			appHubType); err != nil {
			logger.Warn("Discovered Service not found, perhaps already registered")
		}
		if discoveredName != "" {
			labelValue := asset.GetLabels()[labelKey]
			if _, err = getOrCreateAppHubApplication(managementProject, region, labelValue, data); err != nil {
				return fmt.Errorf("error creating application: %w", err)
			}
			displayName := asset.Name[strings.LastIndex(asset.Name, "/")+1:]
			if err = registerServiceWithApplication(managementProject,
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
