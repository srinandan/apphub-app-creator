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
			name:      "DaemonSet should be a workload",
			assetType: "apps.k8s.io/DaemonSet",
			want:      "discoveredWorkload",
		},
		{
			name:      "StatefulSet should be a workload",
			assetType: "apps.k8s.io/StatefulSet",
			want:      "discoveredWorkload",
		},
		{
			name:      "CronJob should be a workload",
			assetType: "batch.k8s.io/CronJob",
			want:      "discoveredWorkload",
		},
		{
			name:      "Cloud Run Job should be a workload",
			assetType: "run.googleapis.com/Job",
			want:      "discoveredWorkload",
		},
		{
			name:      "Cloud Run WorkerPool should be a workload",
			assetType: "run.googleapis.com/WorkerPool",
			want:      "discoveredWorkload",
		},
		{
			name:      "Compute InstanceGroup should be a workload",
			assetType: "compute.googleapis.com/InstanceGroup",
			want:      "discoveredWorkload",
		},
		{
			name:      "ReasoningEngine should be a workload",
			assetType: "aiplatform.googleapis.com/ReasoningEngine",
			want:      "discoveredWorkload",
		},
		{
			name:      "BatchPredictionJob should be a workload",
			assetType: "aiplatform.googleapis.com/BatchPredictionJob",
			want:      "discoveredWorkload",
		},
		{
			name:      "TuningJob should be a workload",
			assetType: "aiplatform.googleapis.com/TuningJob",
			want:      "discoveredWorkload",
		},
		{
			name:      "Cloud Build WorkerPool should be a workload",
			assetType: "cloudbuild.googleapis.com/WorkerPool",
			want:      "discoveredWorkload",
		},
		{
			name:      "Cloud Scheduler Job should be a workload",
			assetType: "cloudscheduler.googleapis.com/Job",
			want:      "discoveredWorkload",
		},
		{
			name:      "Config Deployment should be a workload",
			assetType: "config.googleapis.com/Deployment",
			want:      "discoveredWorkload",
		},
		{
			name:      "Discovery Engine Agent should be a workload",
			assetType: "discoveryengine.googleapis.com/Agent",
			want:      "discoveredWorkload",
		},
		{
			name:      "Discovery Engine Assistant should be a workload",
			assetType: "discoveryengine.googleapis.com/Assistant",
			want:      "discoveredWorkload",
		},
		{
			name:      "Gemini Data Analytics DataAgent should be a workload",
			assetType: "geminidataanalytics.googleapis.com/DataAgent",
			want:      "discoveredWorkload",
		},
		{
			name:      "Cloud Run Service should be a service",
			assetType: "run.googleapis.com/Service",
			want:      "discoveredService",
		},
		{
			name:      "ForwardingRule should be a service",
			assetType: "compute.googleapis.com/ForwardingRule",
			want:      "discoveredService",
		},
		{
			name:      "Storage Bucket should be a service",
			assetType: "storage.googleapis.com/Bucket",
			want:      "discoveredService",
		},
		{
			name:      "Agent Registry McpServer should be a service",
			assetType: "agentregistry.googleapis.com/GoogleMcpServer",
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
				t.Errorf("identifyServiceOrWorkload(%q) = %v, want %v", tt.assetType, got, tt.want)
			}
		})
	}
}
