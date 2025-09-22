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
	"github.com/spf13/cobra"
)

// Cmd to manage apps
var Cmd = &cobra.Command{
	Use:     "apps",
	Aliases: []string{"applications"},
	Short:   "Manage App Hub Applications",
	Long:    "Manage App Hub Applications",
}

var project, managementProject string
var locations []string

func init() {
	Cmd.PersistentFlags().StringVarP(&project, "project", "",
		"", "GCP Project name for CAIS Asset Search")
	Cmd.PersistentFlags().StringArrayVarP(&locations, "locations", "",
		[]string{}, "GCP location names to filter CAIS Asset Search (e.g. us-central1)")
	Cmd.PersistentFlags().StringVarP(&managementProject, "management-project", "",
		"", "App Hub Management Project Id; defaults to project")

	Cmd.AddCommand(GenAppsCmd)
}
