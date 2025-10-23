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
	"regexp"

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
		appName := GetStringParam(cmd.Flag("app-name"))

		if parent == "" {
			return fmt.Errorf("parent is a required field")
		}

		if !IsValidResourceFormat(parent) {
			return fmt.Errorf("parent must be of the format projects/{project} or folders/{folder}")
		}

		if managementProject == "" && IsFolder(parent) {
			return fmt.Errorf("management-project is a required field for folders")
		}

		if len(locations) == 0 {
			return fmt.Errorf("at least one location is required")
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
		if IsFolder(parent) && logLabelKey != "" {
			return fmt.Errorf("log-label-key is not allowed for folders")
		}

		if !isValidAppName(appName) {
			return fmt.Errorf("app-name must start with a lowercase letter")
		}

		return
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		cmd.SilenceUsage = true

		labelKey := GetStringParam(cmd.Flag("label-key"))
		labelValue := GetStringParam(cmd.Flag("label-value"))
		logLabelKey := GetStringParam(cmd.Flag("log-label-key"))
		logLabelValue := GetStringParam(cmd.Flag("log-label-value"))
		tagKey := GetStringParam(cmd.Flag("tag-key"))
		tagValue := GetStringParam(cmd.Flag("tag-value"))
		attributes := GetStringParam(cmd.Flag("attributes"))
		assetTypes := GetStringParam(cmd.Flag("asset-types"))
		contains := GetStringParam(cmd.Flag("contains"))
		appName := GetStringParam(cmd.Flag("app-name"))
		perK8sNamespace, _ := cmd.Flags().GetBool("per-k8s-namespace")
		perK8sAppLabel, _ := cmd.Flags().GetBool("per-k8s-app-label")
		reportOnly, _ := cmd.Flags().GetBool("report-only")
		autoDetect, _ := cmd.Flags().GetBool("auto-detect")

		var attributesData, assetTypesData []byte
		var generatedApplications map[string][]string

		if managementProject == "" {
			managementProject, err = GetProjectID(parent)
			if err != nil {
				return err
			}
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

		if autoDetect {
			generatedApplications, err = client.GenerateFromAll(parent,
				managementProject,
				locations,
				attributesData,
				reportOnly)
		} else if perK8sNamespace {
			generatedApplications, err = client.GenerateAppsPerNamespace(parent,
				managementProject,
				locations,
				attributesData,
				reportOnly)
		} else if perK8sAppLabel {
			generatedApplications, err = client.GenerateKubernetesApps(parent,
				managementProject,
				locations,
				attributesData,
				reportOnly)
		} else if logLabelKey != "" {
			logProject, _ := GetProjectID(parent)
			generatedApplications, err = client.GenerateAppsCloudLogging(logProject,
				managementProject,
				logLabelKey,
				logLabelValue,
				locations,
				attributesData,
				reportOnly)
		} else if len(projectKeys) > 0 {
			generatedApplications, err = client.GenerateFromProject(parent,
				managementProject,
				appName,
				projectKeys,
				locations,
				attributesData,
				nil,
				reportOnly)
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

			if labelValue == "" {
				labelValue = "*"
			}

			generatedApplications, err = client.GenerateAppsAssetInventory(parent,
				managementProject,
				labelKey,
				labelValue,
				tagKey,
				tagValue,
				contains,
				locations,
				attributesData,
				assetTypesData,
				reportOnly)
		}
		if err != nil {
			return err
		}
		if reportOnly {
			PrintGeneratedApplication(generatedApplications)
		}
		return nil
	},
	Example: `Create apps by searching CAIS based on GCP Resource labels in the following locations: ` + genAppsCmdExamples[0] + `

Create apps by searching CAIS based on GCP Resource tags in the following locations: ` + genAppsCmdExamples[1] + `

Create an application per Kubernetes namespace per GKE Cluster in the following locations: ` + genAppsCmdExamples[2] + `

Create apps by searching Cloud Logging labels in the following locations: ` + genAppsCmdExamples[3] + `

Create one App Hub application per app.kubernetes.io/name label value: ` + genAppsCmdExamples[4] + `

Generate a report of discovered assets: ` + genAppsCmdExamples[5] + `

Automatically detect applications based on well known labels and tags: ` + genAppsCmdExamples[6] + `

Generate an application per project or list of projects: ` + genAppsCmdExamples[7],
}

var genAppsCmdExamples = []string{
	`apphub-app-creator apps generate --parent projects/$project --management-project $mp --locations us-west1 --locations us-east1 --label-key $label_key`,
	`apphub-app-creator apps generate --parent projects/$project --management-project $mp --locations us-west1 --label-key $tag_key --tag-value $tag_value`,
	`apphub-app-creator apps generate --parent projects/$project --management-project $mp --locations us-west1 --per-k8s-namespace=true`,
	`apphub-app-creator apps generate --parent folders/$folder --management-project $mp --locations us-west1 --log-label-key $log_label_key --log-label-value $log_label_value`,
	`apphub-app-creator apps generate --parent projects/$project --management-project $mp --locations us-west1 --per-k8s-app-label=true`,
	`apphub-app-creator apps generate --parent projects/$project --management-project $mp --locations us-west1 --label-key $label_key --report-only=true`,
	`apphub-app-creator apps generate --parent projects/$project --management-project $mp --locations us-west1 --auto-detect=true --report-only=true`,
	`apphub-app-creator apps generate --parent folders/$folder --management-project $mp --locations us-west1 --project-keys proj1 --project-keys proj2 --app-name my-app`,
}

func isValidAppName(s string) bool {
	pattern := `^[a-z]`
	isValid, _ := regexp.MatchString(pattern, s)
	return isValid
}

func init() {
	var labelKey, labelValue, tagKey, tagValue, contains, logLabelKey, logLabelValue string
	var attributes, assetTypes, appName string
	var perK8sNamespace, perK8sAppLabel, reportOnly, autoDetect bool

	GenAppsCmd.Flags().StringVarP(&labelKey, "label-key", "",
		"", "Key of the GCP resource label to use for grouping assets into applications.")
	GenAppsCmd.Flags().StringVarP(&labelValue, "label-value", "",
		"", "Value of the GCP resource label to filter assets. If specified, only assets with this label value will be processed.")
	GenAppsCmd.Flags().StringVarP(&tagKey, "tag-key", "",
		"", "Key of the GCP resource tag to use for grouping assets into applications.")
	GenAppsCmd.Flags().StringVarP(&tagValue, "tag-value", "",
		"", "Value of the GCP resource tag to filter assets. If specified, only assets with this tag value will be processed.")
	GenAppsCmd.Flags().StringVarP(&logLabelKey, "log-label-key", "",
		"", "Key of the Cloud Logging log entry label to use for discovering assets.")
	GenAppsCmd.Flags().StringVarP(&logLabelValue, "log-label-value", "",
		"", "Value of the Cloud Logging log entry label, which will also be the application name.")
	GenAppsCmd.Flags().StringVarP(&contains, "contains", "",
		"", "A string that asset resource names must contain. This string will also be the application name.")
	GenAppsCmd.Flags().StringArrayVarP(&projectKeys, "project-keys", "",
		[]string{}, "A list of project ids. Should be used in conjunction with parent=folders/{folder}")
	GenAppsCmd.Flags().StringVarP(&appName, "app-name", "",
		"", "A name for the App Hub Application. Should be used in conjunction with project-keys")
	GenAppsCmd.Flags().StringVarP(&attributes, "attributes", "",
		"", "Path to a json file containing App Hub attributes")
	GenAppsCmd.Flags().BoolVarP(&perK8sNamespace, "per-k8s-namespace", "",
		false, "Create one App Hub application per discovered Kubernetes namespace.")
	GenAppsCmd.Flags().BoolVarP(&perK8sAppLabel, "per-k8s-app-label", "",
		false, "Create one App Hub application per app.kubernetes.io/name label value.")
	GenAppsCmd.Flags().StringVarP(&assetTypes, "asset-types", "",
		"", "Path to a CSV file containing CAIS Asset Types")
	GenAppsCmd.Flags().BoolVarP(&reportOnly, "report-only", "",
		false, "Generates a report of discovered assets without creating applications or registering services/workloads.")
	GenAppsCmd.Flags().BoolVarP(&autoDetect, "auto-detect", "",
		false, "Automatically detect applications using well known identifiers through labels and tags.")

	GenAppsCmd.MarkFlagsMutuallyExclusive("auto-detect", "label-key", "tag-key", "contains", "log-label-key", "per-k8s-namespace", "per-k8s-app-label", "project-keys")
	GenAppsCmd.MarkFlagsMutuallyExclusive("label-value", "tag-value")
	GenAppsCmd.MarkFlagsRequiredTogether("project-keys", "app-name")
	GenAppsCmd.MarkFlagsOneRequired("auto-detect", "label-key", "tag-key", "contains", "log-label-key", "per-k8s-namespace", "per-k8s-app-label", "project-keys")
}
