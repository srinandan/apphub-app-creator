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
	"context"
	"encoding/json"
	"fmt"
	"internal/clilog"
	"io"
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"
)

var clilogger = clilog.GetLogger()

// RootCmd to manage apphub-app-creator
var RootCmd = &cobra.Command{
	Use:   "apphub-app-creator",
	Short: "Utility to generate App Hub Applications.",
	Long:  "This command create App Hub Applications from Cloud Asset Inventory.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var level slog.Level

		switch logLevel {
		case "info":
			level = slog.LevelInfo
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		case "off":
			level = -1
		default:
			return fmt.Errorf("invalid log level: %s", logLevel)
		}

		if logLevel == "off" {
			clilog.Init(nil)
		} else {
			clilog.Init(&slog.HandlerOptions{
				AddSource: true,
				Level:     level,
			})
		}

		logger := clilog.GetLogger()
		if !disableCheck {
			latestVersion, _ := getLatestVersion()
			if cmd.Version == "" {
				logger.Debug("apphub-app-creator wasn't built with a valid Version tag.")
			} else if latestVersion != "" && cmd.Version != latestVersion {
				logger.Info("You are using %s, the latest version %s "+
					"is available for download\n", cmd.Version, latestVersion)
			}
		}

		return nil
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		clilogger.Error("Unable to execute ", "error", err.Error())
	}
}

var (
	logLevel     string
	disableCheck bool
)

func init() {
	RootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info",
		"Set the logging level (info, warn, error or off)")

	RootCmd.PersistentFlags().BoolVarP(&disableCheck, "disable-check", "",
		false, "Disable check for newer versions")

	RootCmd.AddCommand(Cmd)
}

// GetRootCmd returns the root of the cobra command-tree.
func GetRootCmd() *cobra.Command {
	return RootCmd
}

func getLatestVersion() (version string, err error) {
	var req *http.Request
	const endpoint = "https://api.github.com/repos/srinandan/apphub-app-creator/releases/latest"
	logger := clilog.GetLogger()

	client := &http.Client{}
	contentType := "application/json"

	ctx := context.Background()
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", contentType)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return "", err
	}

	if result["tag_name"] == "" {
		logger.Debug("Unable to determine latest tag, skipping this information")
		return "", nil
	}
	return fmt.Sprintf("%s", result["tag_name"]), nil
}
