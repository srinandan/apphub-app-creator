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

package cmd

import (
	"encoding/json"
	"fmt"
	"internal/client"
	"internal/clilog"
	"net/http"
	"strings"
)

type GenerateRequest struct {
	Selector Selector `json:"selector"`
	Action   Action   `json:"action"`
	Options  *Options `json:"options"`
	Scope    Scope    `json:"scope"`
}

type Options struct {
	//Attributes string   `json:"attributes"`
	AssetTypes []string `json:"assetTypes"`
}

type Selector struct {
	Label           *KeyValue   `json:"label"`
	Tag             *KeyValue   `json:"tag"`
	LogLabel        *KeyValue   `json:"logLabel"`
	Contains        *string     `json:"contains"`
	ProjectKeys     *ProjectKey `json:"projectKeys"`
	PerK8sNamespace bool        `json:"perK8sNamespace"`
	PerK8sAppLabel  bool        `json:"perK8sAppLabel"`
	AutoDetect      bool        `json:"autoDetect"`
}

type Scope struct {
	Locations         []string `json:"locations"`
	Parent            string   `json:"parent"`
	ManagementProject string   `json:"managementProject"`
}

type Action struct {
	ReportOnly bool `json:"reportOnly"`
}

type ProjectKey struct {
	AppName    string   `json:"appName"`
	ProjectIds []string `json:"projectIds"`
}

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type GenerateResponse struct {
	Applications map[string][]string `json:"applications"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// ValidateExclusive checks that exactly one field in the Selector struct is set.
func (s *Selector) ValidateExclusive() error {

	boolFields := []string{}
	if s.PerK8sNamespace {
		boolFields = append(boolFields, "perK8sNamespace")
	}
	if s.PerK8sAppLabel {
		boolFields = append(boolFields, "perK8sAppLabel")
	}
	if s.AutoDetect {
		boolFields = append(boolFields, "autoDetect")
	}

	if len(boolFields) > 1 {
		return fmt.Errorf("validation failed: only one of perK8sNamespace, perK8sAppLabel, or autoDetect can be true, but %d were: %s", len(boolFields), strings.Join(boolFields, ", "))
	}

	setFields := []string{}

	if s.Label != nil {
		setFields = append(setFields, "label")
		if err := validateKeyValue(*s.Label); err != nil {
			return err
		}
	}
	if s.Tag != nil {
		setFields = append(setFields, "tag")
		if err := validateKeyValue(*s.Tag); err != nil {
			return err
		}
	}
	if s.LogLabel != nil {
		setFields = append(setFields, "logLabel")
		if err := validateKeyValue(*s.LogLabel); err != nil {
			return err
		}
	}
	if s.Contains != nil {
		setFields = append(setFields, "contains")
	}
	if s.ProjectKeys != nil {
		setFields = append(setFields, "projectKeys")
		if len(s.ProjectKeys.ProjectIds) == 0 {
			return fmt.Errorf("there must be at least one project id")
		}
		if s.ProjectKeys.AppName == "" {
			return fmt.Errorf("appName is a required field")
		}
	}

	count := len(setFields)

	if count == 0 {
		return fmt.Errorf("validation failed: exactly one selector field (label, tag, logLabel, contains, projectKeys, perK8sNamespace, perK8sAppLabel, or autoDetect) must be set, but none were")
	}

	if count > 1 {
		return fmt.Errorf("validation failed: only one selector field can be set, but %d were: %s", count, strings.Join(setFields, ", "))
	}

	return nil
}

// Validate checks the entire GenerateRequest for required fields.
func (s *Scope) Validate() error {
	if s.Parent == "" {
		return fmt.Errorf("validation failed: 'parent' is a required field")
	}
	if s.ManagementProject == "" {
		return fmt.Errorf("validation failed: 'managementProject' is a required field")
	}
	if len(s.Locations) == 0 {
		return fmt.Errorf("validation failed: at least one 'location' is required")
	}
	return nil
}

func (gr *GenerateRequest) Validate() error {
	if err := gr.Selector.ValidateExclusive(); err != nil {
		return err
	}
	if err := gr.Scope.Validate(); err != nil {
		return err
	}
	return nil
}

func validateKeyValue(kv KeyValue) error {
	if kv.Key == "" {
		return fmt.Errorf("validation failed: 'key' is a required field")
	}
	return nil
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	logger := clilog.GetLogger()

	var generateReq GenerateRequest
	var generatedApplications map[string][]string

	err := json.NewDecoder(r.Body).Decode(&generateReq)
	if err != nil {
		logger.Warn("Failed to decode request body", "error", err)
		// Send a 400 Bad Request response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request payload"})
		return
	}

	if generateReq.Selector.AutoDetect {
		generatedApplications, err = client.GenerateFromAll(generateReq.Scope.Parent,
			generateReq.Scope.ManagementProject,
			generateReq.Scope.Locations,
			nil,
			generateReq.Action.ReportOnly)
	} else if generateReq.Selector.PerK8sNamespace {
		generatedApplications, err = client.GenerateAppsPerNamespace(generateReq.Scope.Parent,
			generateReq.Scope.ManagementProject,
			generateReq.Scope.Locations,
			nil,
			generateReq.Action.ReportOnly)
	} else if generateReq.Selector.PerK8sAppLabel {
		generatedApplications, err = client.GenerateKubernetesApps(generateReq.Scope.Parent,
			generateReq.Scope.ManagementProject,
			generateReq.Scope.Locations,
			nil,
			generateReq.Action.ReportOnly)
	} else if generateReq.Selector.LogLabel != nil {
		logProject, _ := GetProjectID(parent)
		generatedApplications, err = client.GenerateAppsCloudLogging(logProject,
			generateReq.Scope.ManagementProject,
			generateReq.Selector.LogLabel.Key,
			generateReq.Selector.LogLabel.Value,
			generateReq.Scope.Locations,
			nil,
			generateReq.Action.ReportOnly)
	} else if generateReq.Selector.ProjectKeys != nil {
		generatedApplications, err = client.GenerateFromProject(generateReq.Scope.Parent,
			generateReq.Scope.ManagementProject,
			generateReq.Selector.ProjectKeys.AppName,
			generateReq.Selector.ProjectKeys.ProjectIds,
			generateReq.Scope.Locations,
			nil,
			nil,
			generateReq.Action.ReportOnly)
	} else {
		if generateReq.Selector.Label != nil && generateReq.Selector.Label.Value == "" {
			generateReq.Selector.Label.Value = "*"
		}
		generatedApplications, err = client.GenerateAppsAssetInventory(generateReq.Scope.Parent,
			generateReq.Scope.ManagementProject,
			generateReq.Selector.Label.Key,
			generateReq.Selector.Label.Value,
			generateReq.Selector.Tag.Key,
			generateReq.Selector.Tag.Value,
			*generateReq.Selector.Contains,
			generateReq.Scope.Locations,
			nil,
			nil,
			generateReq.Action.ReportOnly)
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := GenerateResponse{
		Applications: generatedApplications,
	}

	// Use the new helper to send the JSON response with a 200 OK
	respondWithJSON(w, http.StatusOK, response)

}

// respondWithError is a helper function to write a JSON error response
func respondWithError(w http.ResponseWriter, code int, message string) {
	// Refactored to use the new generic helper
	respondWithJSON(w, code, ErrorResponse{Error: message})
}

// respondWithJSON is a helper function to write any data as a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Encode the payload directly to the response writer
	response, err := json.Marshal(payload)
	if err != nil {
		clilog.GetLogger().Error("Failed to marshal JSON response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to marshal JSON response"}`)) // fallback
		return
	}

	// Write status code and response
	w.WriteHeader(code)
	w.Write(response)
}
