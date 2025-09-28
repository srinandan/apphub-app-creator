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

	"cloud.google.com/go/logging"
	"cloud.google.com/go/logging/logadmin"
	"google.golang.org/api/iterator"
)

type logAsset struct {
	Name       string
	AppHubType string
	Location   string
}

var INCLUDED_RESOURCE_TYPES = []string{
	"cloud_run_revision",
	"k8s_pod",
	"gce_instance_group",
}

const k8s_deployment = "AND labels.\"logging.gke.io/top_level_controller_type\"=\"Deployment\""

func filterLogs(projectID, labelKey, labelValue string, locations []string) (map[string]logAsset, error) {
	ctx := context.Background()
	logger := clilog.GetLogger()

	assets := make(map[string]logAsset)

	// Create the Log Admin Client
	client, err := logadmin.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("logadmin.NewClient: %w", err)
	}
	defer client.Close()

	filter := fmt.Sprintf("%s AND (labels.%s=\"%s\") AND %s", generateLocationFilter(locations),
		labelKey, labelValue, generateResourceTypeFilter())

	logger.Info("Searching logs with query", "query", filter)

	// Execute the query using the constructed filter
	it := client.Entries(ctx, logadmin.Filter(filter))

	// Iterate over the results
	for {
		entry, err := it.Next()

		if err == iterator.Done {
			break // No more entries
		}
		if err != nil {
			return nil, fmt.Errorf("it.Next: %w", err)
		}
		asset, l := getAsset(entry)
		if asset != "" {
			assets[asset] = l
		}
	}
	return assets, nil
}

// generateLocationFilter takes a string array of locations (e.g., "us-central1,europe-west1")
// and returns a filter string in the format (resource.location="loc1" OR resource.location="loc2").
func generateLocationFilter(locations []string) string {

	var clauses []string

	for _, loc := range locations {
		clause := fmt.Sprintf(`resource.labels.location="%s"`, loc)
		clauses = append(clauses, clause)
	}

	// Join the clauses with " OR ".
	filter := strings.Join(clauses, " OR ")

	// Enclose the entire expression in parentheses.
	if filter != "" {
		return fmt.Sprintf("(%s)", filter)
	}

	return ""
}

// generateResourceTypeFilter returns a filter string in the format (resource.type="type1" OR resource.type="type2").
func generateResourceTypeFilter() string {

	var clauses []string

	for _, rt := range INCLUDED_RESOURCE_TYPES {
		var clause string
		if rt == "k8s_pod" {
			clause = fmt.Sprintf(`(resource.type="%s" AND %s)`, rt, k8s_deployment)
		} else {
			clause = fmt.Sprintf(`resource.type="%s"`, rt)
		}
		clauses = append(clauses, clause)
	}

	// Join the clauses with " OR ".
	filter := strings.Join(clauses, " OR ")

	// Enclose the entire expression in parentheses.
	if filter != "" {
		return fmt.Sprintf("(%s)", filter)
	}

	return ""
}

func getAsset(entry *logging.Entry) (string, logAsset) {
	switch entry.Resource.Type {
	case "cloud_run_revision":
		return fmt.Sprintf("//run.googleapis.com/projects/%s/locations/%s/services/%s",
				entry.Resource.Labels["project_id"], entry.Resource.Labels["location"],
				entry.Resource.Labels["service_name"]), logAsset{
				Name:       entry.Resource.Labels["service_name"],
				AppHubType: "discoveredService",
				Location:   entry.Resource.Labels["location"],
			}
	default:
		return "", logAsset{}
	}
}
