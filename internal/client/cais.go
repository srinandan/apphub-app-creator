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
	// runtimes
	"run.googleapis.com/Service",
	"run.googleapis.com/Job",
	"apps.k8s.io/Deployment",
	"apps.k8s.io/DaemonSet",
	"apps.k8s.io/StatefulSet",
	"compute.googleapis.com/InstanceGroup",
	// networking
	"compute.googleapis.com/ForwardingRule",
	"compute.googleapis.com/BackendService",
	"gateway.networking.k8s.io/Gateway",
	// storage
	"storage.googleapis.com/Bucket",
	"pubsub.googleapis.com/Topic",
	"pubsub.googleapis.com/Subscription",
	// databases
	"alloydb.googleapis.com/Instance",
	"spanner.googleapis.com/Instance",
	"sqladmin.googleapis.com/Instance",
	"alloydb.googleapis.com/Instance",
	"redis.googleapis.com/Instance",
}

var KUBERNETES_ASSETS = []string{
	"apps.k8s.io/Deployment",
	"apps.k8s.io/DaemonSet",
	"apps.k8s.io/StatefulSet",
	"k8s.io/Service",
	"gateway.networking.k8s.io/Gateway",
}

var MAX_PAGE int32 = 1000

// searchAssets queries the Cloud Asset Inventory for resources within a specific project
// and location
func searchAssets(parent, labelKey, labelValue, tagKey, tagValue, contains string, locations []string, assetTypesData []byte) ([]*assetpb.ResourceSearchResult, error) {
	ctx := context.Background()
	var searchAssetTypes []string
	var queryParts []string

	logger := clilog.GetLogger()
	// Initialize the Asset Service client
	client, err := asset.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset client: %w", err)
	}
	defer client.Close()

	// Build the full search query.
	if len(locations) > 1 {
		queryParts = append(queryParts, fmt.Sprintf("location:(%s)", strings.Join(locations, " OR ")))
	} else {
		queryParts = []string{fmt.Sprintf("location:%s", locations[0])}
	}

	if labelKey != "" {
		if labelValue != "" {
			queryParts = append(queryParts, fmt.Sprintf("labels.%s:%s", labelKey, labelValue))
		} else {
			queryParts = append(queryParts, fmt.Sprintf("labels.%s:*", labelKey))
		}
	} else if tagKey != "" {
		if tagValue != "" {
			queryParts = append(queryParts, fmt.Sprintf("tagKeys.%s:%s", tagKey, tagValue))
		} else {
			queryParts = append(queryParts, fmt.Sprintf("tagKeys.%s:*", tagKey))
		}
	} else if contains != "" {
		queryParts = append(queryParts, fmt.Sprintf("name:%s", contains))
	}

	fullQuery := strings.Join(queryParts, " AND ")

	logger.Info("Searching scope with query", "scope", parent, "query", fullQuery)

	if len(assetTypesData) > 0 {
		searchAssetTypes = strings.Split(string(assetTypesData), ",")
	} else {
		searchAssetTypes = INCLUDED_ASSETS
	}

	logger.Info("Searching asset types", "assets", searchAssetTypes)

	// Construct the search request
	req := &assetpb.SearchAllResourcesRequest{
		Scope:      parent,
		Query:      fullQuery,
		AssetTypes: searchAssetTypes,
		PageSize:   MAX_PAGE,
	}

	// Call SearchAllResources and iterate over the results
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

// searchKubernetes queries the Cloud Asset Inventory for kubernetes resources within a specific project
// and location
func searchKubernetes(parent string, locations []string) ([]*assetpb.ResourceSearchResult, error) {
	ctx := context.Background()
	var searchAssetTypes []string
	var queryParts []string

	logger := clilog.GetLogger()
	// Initialize the Asset Service client
	client, err := asset.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset client: %w", err)
	}
	defer client.Close()

	// Build the full search query.
	if len(locations) > 1 {
		queryParts = append(queryParts, fmt.Sprintf("location:(%s)", strings.Join(locations, " OR ")))
	} else {
		queryParts = []string{fmt.Sprintf("location:%s", locations[0])}
	}

	// exclude kubernetes system namespaces
	queryParts = append(queryParts, "NOT parentFullResourceName : \"kube-system\" AND "+
		"NOT parentFullResourceName : \"gmp-system\" AND NOT parentFullResourceName : \"gke-managed-cim\"")

	fullQuery := strings.Join(queryParts, " AND ")

	logger.Info("Searching scope with query", "scope", parent, "query", fullQuery)

	searchAssetTypes = KUBERNETES_ASSETS

	logger.Info("Searching asset types", "assets", searchAssetTypes)

	// Construct the search request
	req := &assetpb.SearchAllResourcesRequest{
		Scope:      parent,
		Query:      fullQuery,
		AssetTypes: searchAssetTypes,
		PageSize:   MAX_PAGE,
	}

	// Call SearchAllResources and iterate over the results
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
