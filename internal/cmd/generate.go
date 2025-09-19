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
	"fmt"
	"internal/client"
	"os"

	"github.com/spf13/cobra"
)

// GenAppsCmd to generate applications
var GenAppsCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate App Hub Applications",
	Long:  "Generate App Hub Applications based on CAIS Asset Search",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if project == "" {
			return fmt.Errorf("project-id is a required field")
		}
		if region == "" {
			return fmt.Errorf("region is a required field")
		}
		return
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		labelKey := GetStringParam(cmd.Flag("label-key"))
		attributes := GetStringParam(cmd.Flag("attributes"))

		var data []byte

		if managementProject == "" {
			managementProject = project
		}

		if attributes != "" {
			if _, err := os.Stat(attributes); os.IsNotExist(err) {
				return err
			}

			data, err = os.ReadFile(attributes)
			if err != nil {
				return err
			}
		}

		err = client.GenerateAppsByLabel(project, managementProject, region, labelKey, data)
		if err != nil {
			return
		}
		return
	},
}

func init() {
	var labelKey, attributes string

	GenAppsCmd.Flags().StringVarP(&labelKey, "label-key", "l",
		"", "GCP Resource Label Key to filter CAIS Resource")
	GenAppsCmd.Flags().StringVarP(&attributes, "attributes", "a",
		"", "Path to a json file containing App Hub attributes")

	_ = GenAppsCmd.MarkFlagRequired("label-key")
}
