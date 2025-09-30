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
	"internal/clilog"

	"github.com/spf13/cobra"
)

var clilogger = clilog.GetLogger()

// RootCmd to manage apphub-app-creator
var RootCmd = &cobra.Command{
	Use:   "apphub-app-creator",
	Short: "Utility to generate App Hub Applications.",
	Long:  "This command create App Hub Applications from Cloud Asset Inventory.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		clilog.Init()
		return nil
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		clilogger.Error("Unable to execute ", "error", err.Error())
	}
}

func init() {
	RootCmd.AddCommand(Cmd)
}

// GetRootCmd returns the root of the cobra command-tree.
func GetRootCmd() *cobra.Command {
	return RootCmd
}
