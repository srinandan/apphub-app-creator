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
		labelKey := GetStringParam(cmd.Flag("label-key"))
		labelValue := GetStringParam(cmd.Flag("label-value"))
		logLabelKey := GetStringParam(cmd.Flag("log-label-key"))
		logLabelValue := GetStringParam(cmd.Flag("log-label-value"))
		tagKey := GetStringParam(cmd.Flag("tag-key"))
		tagValue := GetStringParam(cmd.Flag("tag-value"))
		contains := GetStringParam(cmd.Flag("contains"))

		if project == "" {
			return fmt.Errorf("project is a required field")
		}
		if len(locations) == 0 {
			return fmt.Errorf("at least one location is required")
		}
		if labelKey == "" && tagKey == "" && contains == "" && logLabelKey == "" {
			return fmt.Errorf("label-key or tag-key or contains or log-label-key is a required field")
		}
		if (labelKey != "" && tagKey != "") ||
			(labelKey != "" && contains != "") ||
			(tagKey != "" && contains != "") ||
			(labelKey != "" && logLabelKey != "") ||
			(tagKey != "" && logLabelKey != "") ||
			(contains != "" && logLabelKey != "") {
			return fmt.Errorf("only one of label-key, tag-key, log-label-key or contains is allowed")
		}
		if labelValue != "" && labelKey == "" {
			return fmt.Errorf("label-value must be used with label-key")
		}
		if tagValue != "" && tagKey == "" {
			return fmt.Errorf("tag-value must be used with tag-key")
		}
		if logLabelKey != "" && logLabelValue == "" {
			return fmt.Errorf("log-label-value must be used with log-label-key")
		}
		return
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		labelKey := GetStringParam(cmd.Flag("label-key"))
		labelValue := GetStringParam(cmd.Flag("label-value"))
		logLabelKey := GetStringParam(cmd.Flag("log-label-key"))
		logLabelValue := GetStringParam(cmd.Flag("log-label-value"))
		tagKey := GetStringParam(cmd.Flag("tag-key"))
		tagValue := GetStringParam(cmd.Flag("tag-value"))
		attributes := GetStringParam(cmd.Flag("attributes"))
		assetTypes := GetStringParam(cmd.Flag("asset-types"))
		contains := GetStringParam(cmd.Flag("contains"))
		var attributesData, assetTypesData []byte

		if managementProject == "" {
			managementProject = project
		}

		if attributes != "" {
			if _, err := os.Stat(attributes); os.IsNotExist(err) {
				return err
			}

			attributesData, err = os.ReadFile(attributes)
			if err != nil {
				return err
			}
		}

		if logLabelKey != "" {
			err = client.GenerateAppsCloudLogging(project,
				managementProject,
				logLabelKey,
				logLabelValue,
				locations,
				attributesData)

			return err
		} else {
			if assetTypes != "" {
				if _, err := os.Stat(assetTypes); os.IsNotExist(err) {
					return err
				}

				assetTypesData, err = os.ReadFile(assetTypes)
				if err != nil {
					return err
				}
			}

			err = client.GenerateAppsAssetInventory(project,
				managementProject,
				labelKey,
				labelValue,
				tagKey,
				tagValue,
				contains,
				locations,
				attributesData,
				assetTypesData)

			return err
		}
	},
}

func init() {
	var labelKey, labelValue, tagKey, tagValue, contains, logLabelKey, logLabelValue string
	var attributes, assetTypes string

	GenAppsCmd.Flags().StringVarP(&labelKey, "label-key", "",
		"", "GCP Resource Label Key to filter CAIS Resource")
	GenAppsCmd.Flags().StringVarP(&labelValue, "label-value", "",
		"", "GCP Resource Label Value to filter CAIS Resource; Must be used with label-key")
	GenAppsCmd.Flags().StringVarP(&tagKey, "tag-key", "",
		"", "GCP Resource Tag Key to filter CAIS Resource")
	GenAppsCmd.Flags().StringVarP(&tagValue, "tag-value", "",
		"", "GCP Resource Tag Value to filter CAIS Resource; Must be used with tag-key")
	GenAppsCmd.Flags().StringVarP(&logLabelKey, "log-label-key", "",
		"", "GCP Cloud Logging Label Key to filter")
	GenAppsCmd.Flags().StringVarP(&logLabelValue, "log-label-value", "",
		"", "GCP Cloud Logging Label Value to filter; Must be used with log-label-key")
	GenAppsCmd.Flags().StringVarP(&contains, "contains", "",
		"", "GCP Resources whose name contains the string")
	GenAppsCmd.Flags().StringVarP(&attributes, "attributes", "",
		"", "Path to a json file containing App Hub attributes")
	GenAppsCmd.Flags().StringVarP(&assetTypes, "asset-types", "",
		"", "Path to a CSV file containing CAIS Asset Types")
}
