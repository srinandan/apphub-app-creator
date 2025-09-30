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
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	apphubpb "cloud.google.com/go/apphub/apiv1/apphubpb"
)

func TestNewAttributesFromBytes(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    *apphubpb.Attributes
		wantErr bool
	}{
		{
			name: "Valid JSON",
			data: []byte(`{"criticality":{"type":"MISSION_CRITICAL"},"environment":{"type":"PRODUCTION"}}`),
			want: &apphubpb.Attributes{
				Criticality: &apphubpb.Criticality{
					Type: apphubpb.Criticality_MISSION_CRITICAL,
				},
				Environment: &apphubpb.Environment{
					Type: apphubpb.Environment_PRODUCTION,
				},
			},
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			data:    []byte(`{"criticality":{"type":"MISSION_CRITICAL"}`),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Empty data",
			data:    []byte{},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newAttributesFromBytes(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("newAttributesFromBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf(`newAttributesFromBytes() mismatch (-want +got):
%s`, diff)
			}
		})
	}
}
