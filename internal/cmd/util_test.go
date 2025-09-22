
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
	"testing"

	"github.com/spf13/pflag"
)

func TestGetStringParam(t *testing.T) {
	tests := []struct {
		name     string
		flag     *pflag.Flag
		expected string
	}{
		{
			name:     "nil flag",
			flag:     nil,
			expected: "",
		},
		{
			name: "flag with value",
			flag: &pflag.Flag{
				Value: func() pflag.Value {
					var v pflag.Value
					v = new(stringValue)
					v.Set("test")
					return v
				}(),
			},
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := GetStringParam(tt.flag)
			if actual != tt.expected {
				t.Errorf("GetStringParam() = %v, want %v", actual, tt.expected)
			}
		})
	}
}

type stringValue string

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Type() string {
	return "string"
}

func (s *stringValue) String() string {
	return string(*s)
}
