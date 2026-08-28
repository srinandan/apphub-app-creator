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
	"bytes"
	"testing"
)

func TestQueryTracesByLabelValidation(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		filter    string
		wantErr   bool
	}{
		{
			name:      "missing projectID and filter",
			projectID: "",
			filter:    "",
			wantErr:   true,
		},
		{
			name:      "missing projectID",
			projectID: "",
			filter:    "label_key:label_val",
			wantErr:   true,
		},
		{
			name:      "missing filter",
			projectID: "my-proj",
			filter:    "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := queryTracesByLabel(&buf, tt.projectID, tt.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("queryTracesByLabel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
