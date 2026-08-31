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
	"log/slog"
	"slices"
	"strings"

	asset "cloud.google.com/go/asset/apiv1"
	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

var INCLUDED_ASSETS = []string{
	// runtimes
	"run.googleapis.com/Service",
	"run.googleapis.com/Job",
	"apps.k8s.io/Deployment",
	"apps.k8s.io/DaemonSet",
	"apps.k8s.io/StatefulSet",
	"compute.googleapis.com/InstanceGroup",
	"aiplatform.googleapis.com/ReasoningEngine",
	"cloudfunctions.googleapis.com/CloudFunction",
	// networking
	"compute.googleapis.com/ForwardingRule",
	"compute.googleapis.com/BackendService",
	"compute.googleapis.com/ServiceAttachment",
	// storage & messaging
	"storage.googleapis.com/Bucket",
	"pubsub.googleapis.com/Topic",
	"pubsub.googleapis.com/Subscription",
	// databases & analytics
	"alloydb.googleapis.com/Instance",
	"spanner.googleapis.com/Instance",
	"spanner.googleapis.com/Database",
	"sqladmin.googleapis.com/Instance",
	"redis.googleapis.com/Instance",
	"bigquery.googleapis.com/Dataset",
	"firestore.googleapis.com/Database",
	// AI & config & workflows
	"aiplatform.googleapis.com/Endpoint",
	"secretmanager.googleapis.com/Secret",
	"workflows.googleapis.com/Workflow",
	"container.googleapis.com/Cluster",
}

var KUBERNETES_ASSETS = []string{
	"apps.k8s.io/Deployment",
	"apps.k8s.io/DaemonSet",
	"apps.k8s.io/StatefulSet",
	"batch.k8s.io/CronJob",
	"k8s.io/Service",
	"gateway.networking.k8s.io/Gateway",
	"networking.k8s.io/Ingress",
}

var WORKLOADS = []string{
	"aiplatform.googleapis.com/BatchPredictionJob",
	"aiplatform.googleapis.com/ReasoningEngine",
	"aiplatform.googleapis.com/TuningJob",
	"apps.k8s.io/DaemonSet",
	"apps.k8s.io/Deployment",
	"apps.k8s.io/StatefulSet",
	"batch.k8s.io/CronJob",
	"cloudbuild.googleapis.com/WorkerPool",
	"cloudscheduler.googleapis.com/Job",
	"compute.googleapis.com/InstanceGroup",
	"config.googleapis.com/Deployment",
	"discoveryengine.googleapis.com/Agent",
	"discoveryengine.googleapis.com/Assistant",
	"geminidataanalytics.googleapis.com/DataAgent",
	"run.googleapis.com/Job",
	"run.googleapis.com/WorkerPool",
}

var GKE_EXCLUSION_NAMESPACES = []string{
	// Core Kubernetes system namespaces
	"kube-system",
	"kube-public",
	"kube-node-lease",
	// GKE-managed & addon system namespaces
	"gke-gmp-system",
	"gke-connect",
	"gke-backup",
	"gke-managed-cim",
	"gke-managed-filestorecsi",
	"gke-managed-metrics-server",
	"gke-managed-system",
	"gke-managed-volumepopulator",
	"gke-managed-dpv2-operator",
	"gke-system",
	"gmp-public",
	"gmp-system",
	"asm-system",
	"istio-system",
	"gatekeeper-system",
	"config-management-system",
	"anthos-identity-service",
	"cert-manager",
}

var MAX_PAGE int32 = 1000

const K8S_APP_LABEL = "app.kubernetes.io/name"

func getExcludedNamespacesFilter() string {
	var gkeExlNs []string
	for _, ns := range GKE_EXCLUSION_NAMESPACES {
		gkeExlNs = append(gkeExlNs, fmt.Sprintf("parentFullResourceName : \"%s\"", ns))
	}
	return fmt.Sprintf("NOT (%s)", strings.Join(gkeExlNs, " OR "))
}

func isExcludedNamespace(parentFullResourceName string) bool {
	if parentFullResourceName == "" {
		return false
	}

	// Extract the namespace name from the parentFullResourceName
	ns := parentFullResourceName
	if idx := strings.LastIndex(parentFullResourceName, "/namespaces/"); idx != -1 {
		ns = parentFullResourceName[idx+len("/namespaces/"):]
		if slashIdx := strings.Index(ns, "/"); slashIdx != -1 {
			ns = ns[:slashIdx]
		}
	} else if idx := strings.LastIndex(parentFullResourceName, "/"); idx != -1 {
		ns = parentFullResourceName[idx+1:]
	}

	if ns == "default" {
		return false
	}

	if strings.HasPrefix(ns, "kube-") || strings.HasPrefix(ns, "gke-") {
		return true
	}

	return slices.Contains(GKE_EXCLUSION_NAMESPACES, ns)
}

// searchAssets queries the Cloud Asset Inventory for resources within a specific project
// and location
func searchAssets(parent, labelKey, labelValue, tagKey, tagValue, contains string, locations []string, assetTypesData []byte) ([]*assetpb.ResourceSearchResult, error) {
	ctx := context.Background()
	var searchAssetTypes []string
	var queryParts []string

	logger := slog.Default()
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
			queryParts = append(queryParts, fmt.Sprintf("labels:%s", labelKey))
		}
	} else if tagKey != "" {
		if tagValue != "" {
			queryParts = append(queryParts,
				fmt.Sprintf("((tagKeys:%s AND tagValues:%s) OR (effectiveTagKeys:%s AND effectiveTagValues:%s))",
					tagKey, tagValue, tagKey, tagValue))
		} else {
			queryParts = append(queryParts, fmt.Sprintf("(tagKeys:%s OR effectiveTagKeys:%s)", tagKey, tagKey))
		}
	} else if contains != "" {
		queryParts = append(queryParts, fmt.Sprintf("name:%s", contains))
	}

	// exclude kubernetes system namespaces
	queryParts = append(queryParts, getExcludedNamespacesFilter())

	fullQuery := strings.Join(queryParts, " AND ")

	logger.Info("Searching scope with query", "scope", parent, "query", fullQuery)

	if len(assetTypesData) > 0 {
		searchAssetTypes = strings.Split(string(assetTypesData), ",")
	} else {
		searchAssetTypes = INCLUDED_ASSETS
	}

	logger.Info("Searching asset types", "assets", searchAssetTypes)

	readMask, _ := fieldmaskpb.New(&assetpb.ResourceSearchResult{}, "*")

	// Construct the search request
	req := &assetpb.SearchAllResourcesRequest{
		Scope:      parent,
		Query:      fullQuery,
		AssetTypes: searchAssetTypes,
		PageSize:   MAX_PAGE,
		ReadMask:   readMask,
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

	logger := slog.Default()
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
	queryParts = append(queryParts, getExcludedNamespacesFilter())

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
		OrderBy:    "parentFullResourceName",
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

// searchKubernetesApps queries the Cloud Asset Inventory for kubernetes resources
// that matches a specific label within a specific project and location
func searchKubernetesApps(parent string, locations []string) ([]*assetpb.ResourceSearchResult, error) {
	ctx := context.Background()
	var searchAssetTypes []string
	var queryParts []string

	logger := slog.Default()
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

	// include kubernetes app label
	queryParts = append(queryParts, fmt.Sprintf("labels.\"%s\":*", K8S_APP_LABEL))

	// exclude kubernetes system namespaces
	queryParts = append(queryParts, getExcludedNamespacesFilter())

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

func searchProject(parent string, projectIds, locations []string, assetTypesData []byte) ([]*assetpb.ResourceSearchResult, error) {
	ctx := context.Background()
	var searchAssetTypes []string
	var queryParts []string

	logger := slog.Default()
	// Initialize the Asset Service client
	client, err := asset.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create asset client: %w", err)
	}
	defer client.Close()

	// Build the full search query.
	if len(locations) > 1 {
		var loc []string
		for _, l := range locations {
			loc = append(loc, fmt.Sprintf("location:%s", l))
		}
		queryParts = append(queryParts, fmt.Sprintf("(%s)", strings.Join(loc, " OR ")))
	} else {
		queryParts = []string{fmt.Sprintf("location:%s", locations[0])}
	}

	// exclude kubernetes system namespaces
	queryParts = append(queryParts, getExcludedNamespacesFilter())

	if len(projectIds) > 1 {
		var p []string
		for _, i := range projectIds {
			p = append(p, fmt.Sprintf("projects/%s", i))
		}
		queryParts = append(queryParts, fmt.Sprintf("(%s)", strings.Join(p, " OR ")))
	}

	fullQuery := strings.Join(queryParts, " AND ")

	logger.Info("Searching scope with query", "scope", parent, "query", fullQuery)

	if len(assetTypesData) > 0 {
		searchAssetTypes = strings.Split(string(assetTypesData), ",")
	} else {
		searchAssetTypes = INCLUDED_ASSETS
	}

	logger.Info("Searching asset types", "assets", searchAssetTypes)

	readMask, _ := fieldmaskpb.New(&assetpb.ResourceSearchResult{}, "*")

	// Construct the search request
	req := &assetpb.SearchAllResourcesRequest{
		Scope:      parent,
		Query:      fullQuery,
		AssetTypes: searchAssetTypes,
		PageSize:   MAX_PAGE,
		ReadMask:   readMask,
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
	if slices.Contains(WORKLOADS, assetType) {
		return "discoveredWorkload"
	}
	return "discoveredService"
}
