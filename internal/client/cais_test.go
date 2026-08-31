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
	"strings"
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

func TestGetExcludedNamespacesFilter(t *testing.T) {
	filter := getExcludedNamespacesFilter()
	if filter == "" {
		t.Fatalf("expected non-empty excluded namespaces filter")
	}

	for _, ns := range GKE_EXCLUSION_NAMESPACES {
		expectedSnippet := "parentFullResourceName : \"" + ns + "\""
		if !strings.Contains(filter, expectedSnippet) {
			t.Errorf("expected filter %q to contain %q", filter, expectedSnippet)
		}
	}
}

func TestIsExcludedNamespace(t *testing.T) {
	tests := []struct {
		name                   string
		parentFullResourceName string
		want                   bool
	}{
		{
			name:                   "empty parent",
			parentFullResourceName: "",
			want:                   false,
		},
		{
			name:                   "standard user namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/default",
			want:                   false,
		},
		{
			name:                   "custom application namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/frontend-prod",
			want:                   false,
		},
		{
			name:                   "kube-system namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/kube-system",
			want:                   true,
		},
		{
			name:                   "kube-public namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/kube-public",
			want:                   true,
		},
		{
			name:                   "kube-node-lease namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/kube-node-lease",
			want:                   true,
		},
		{
			name:                   "gke-managed-system namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/gke-managed-system",
			want:                   true,
		},
		{
			name:                   "gke-system namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/gke-system",
			want:                   true,
		},
		{
			name:                   "gmp-system namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/gmp-system",
			want:                   true,
		},
		{
			name:                   "gke-backup namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/gke-backup",
			want:                   true,
		},
		{
			name:                   "gke-connect namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/gke-connect",
			want:                   true,
		},
		{
			name:                   "gke-managed-metrics-server namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/gke-managed-metrics-server",
			want:                   true,
		},
		{
			name:                   "gke-managed-filestorecsi namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/gke-managed-filestorecsi",
			want:                   true,
		},
		{
			name:                   "gke-managed-volumepopulator namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/gke-managed-volumepopulator",
			want:                   true,
		},
		{
			name:                   "gke-managed-dpv2-operator namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/gke-managed-dpv2-operator",
			want:                   true,
		},
		{
			name:                   "gmp-public namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/gmp-public",
			want:                   true,
		},
		{
			name:                   "asm-system namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/asm-system",
			want:                   true,
		},
		{
			name:                   "gatekeeper-system namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/gatekeeper-system",
			want:                   true,
		},
		{
			name:                   "config-management-system namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/config-management-system",
			want:                   true,
		},
		{
			name:                   "anthos-identity-service namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/anthos-identity-service",
			want:                   true,
		},
		{
			name:                   "cert-manager namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/cert-manager",
			want:                   true,
		},
		{
			name:                   "custom kube- prefix namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/kube-custom",
			want:                   true,
		},
		{
			name:                   "custom gke- prefix namespace",
			parentFullResourceName: "//container.googleapis.com/projects/p1/locations/us-central1/clusters/c1/k8s/namespaces/gke-custom",
			want:                   true,
		},
		{
			name:                   "bare namespace name kube-system",
			parentFullResourceName: "kube-system",
			want:                   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isExcludedNamespace(tt.parentFullResourceName)
			if got != tt.want {
				t.Errorf("isExcludedNamespace(%q) = %v, want %v", tt.parentFullResourceName, got, tt.want)
			}
		})
	}
}
