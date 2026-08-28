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

package clilog

import (
	"log/slog"
	"testing"
)

func TestInitAndGetLogger(t *testing.T) {
	tests := []struct {
		name string
		opts *slog.HandlerOptions
	}{
		{
			name: "Init with nil options discards logs",
			opts: nil,
		},
		{
			name: "Init with custom handler options",
			opts: &slog.HandlerOptions{
				Level:     slog.LevelDebug,
				AddSource: true,
			},
		},
		{
			name: "Init with info level",
			opts: &slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.opts)
			l := GetLogger()
			if l == nil {
				t.Fatalf("GetLogger() returned nil after Init(%v)", tt.opts)
			}
			// Verify logger is callable
			l.Info("test message", "key", "value")
		})
	}
}
