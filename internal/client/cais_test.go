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
	"testing"
)

func TestIdentifyServiceOrWorkload(t *testing.T) {
	tests := []struct {
		name      string
		assetType string
		want      string
	}{
		{
			name:      "Deployment should be a workload",
			assetType: "apps.k8s.io/Deployment",
			want:      "discoveredWorkload",
		},
		{
			name:      "Cloud Run Service should be a service",
			assetType: "run.googleapis.com/Service",
			want:      "discoveredService",
		},
		{
			name:      "Unknown type should be a service",
			assetType: "some.other.asset/Type",
			want:      "discoveredService",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := identifyServiceOrWorkload(tt.assetType); got != tt.want {
				t.Errorf("identifyServiceOrWorkload() = %v, want %v", got, tt.want)
			}
		})
	}
}
