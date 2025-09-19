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
	"slices"
	"strings"

	asset "cloud.google.com/go/asset/apiv1"
	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	"google.golang.org/api/iterator"
)

var INCLUDED_ASSETS = []string{
	"run.googleapis.com/Service",
	"sqladmin.googleapis.com/Instance",
	"storage.googleapis.com/Bucket",
	"spanner.googleapis.com/Instance",
	"run.googleapis.com/Job",
}

// searchAssets queries the Cloud Asset Inventory for resources within a specific project
// and region, optionally applying an additional query filter.
//
// The region filter is implemented using the CAIS Query Language: "location:{region}".
func searchAssets(projectID, region, labelKey string) ([]*assetpb.ResourceSearchResult, error) {
	ctx := context.Background()

	logger := clilog.GetLogger()
	// 1. Initialize the Asset Service client
	// The client automatically picks up credentials from the environment (e.g., Application Default Credentials).
	client, err := asset.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset client: %w", err)
	}
	defer client.Close()

	// 2. Construct the search scope and query
	// The scope is the resource under which to search (Project, Folder, or Organization).
	scope := fmt.Sprintf("projects/%s", projectID)

	// Build the full search query: location filter is mandatory, additional query is optional.
	queryParts := []string{fmt.Sprintf("location:%s", region), fmt.Sprintf("labels.%s:*", labelKey)}

	fullQuery := strings.Join(queryParts, " AND ")

	logger.Info("Searching scope with query", "scope", scope, "query", fullQuery)

	// 3. Construct the search request
	req := &assetpb.SearchAllResourcesRequest{
		Scope:      scope,
		Query:      fullQuery,
		AssetTypes: INCLUDED_ASSETS,
	}

	// 4. Call SearchAllResources and iterate over the results
	var assets []*assetpb.ResourceSearchResult
	it := client.SearchAllResources(ctx, req)

	for {
		asset, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error while iterating resources: %w", err)
		}
		assets = append(assets, asset)
	}

	return assets, nil
}

func identifyServiceOrWorkload(assetType string) string {
	WORKLOADS := []string{
		"apps.k8s.io/Deployment",
		"apps.k8s.io/DaemonSet",
		"apps.k8s.io/StatefulSet",
		"run.googleapis.com/Job",
		"compute.googleapis.com/InstanceGroup",
	}
	if slices.Contains(WORKLOADS, assetType) {
		return "discoveredWorkload"
	} else {
		return "discoveredService"
	}
}
