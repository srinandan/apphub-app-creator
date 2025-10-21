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
	"io"

	trace "cloud.google.com/go/trace/apiv1"
	"cloud.google.com/go/trace/apiv1/tracepb"
	"google.golang.org/api/iterator"
)

// queryTracesByLabel queries and prints traces that match a given filter.
// The filter string is used to specify which labels to match.
func queryTracesByLabel(w io.Writer, projectID, filter string) error {
	// A filter is required to query traces.
	if projectID == "" || filter == "" {
		return fmt.Errorf("projectID and filter must be specified")
	}

	ctx := context.Background()

	logger := clilog.GetLogger()

	// 1. Create a new Cloud Trace client.
	// This will use Application Default Credentials to authenticate.
	c, err := trace.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create trace client: %w", err)
	}
	defer c.Close()

	// 2. Build the request for the ListTraces API call.
	req := &tracepb.ListTracesRequest{
		// The parent resource is the project ID.
		ProjectId: fmt.Sprintf("projects/%s", projectID),
		// The filter string is the key to querying by labels.
		Filter: filter,
	}

	// 3. Call the ListTraces API.
	// This returns an iterator that we can loop over.
	it := c.ListTraces(ctx, req)
	logger.Info("Traces found for filter", "filter", filter)

	// 4. Iterate over the results.
	for {
		trace, err := it.Next()
		// iterator.Done is returned when there are no more results.
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to retrieve next trace: %w", err)
		}
		logger.Info("Trace ID", "id", trace.TraceId)
	}

	return nil
}
