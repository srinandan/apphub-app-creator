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
	apphubpb "cloud.google.com/go/apphub/apiv1/apphubpb"
	"google.golang.org/protobuf/encoding/protojson"
)

// NewAttributesFromBytes takes a JSON byte array and unmarshals it
// directly into an apphubpb.Attributes struct using protojson.
// This is the idiomatic way to convert JSON to a protobuf message in Go.
func newAttributesFromBytes(data []byte) (*apphubpb.Attributes, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var attributes apphubpb.Attributes
	if err := protojson.Unmarshal(data, &attributes); err != nil {
		return nil, err
	}
	return &attributes, nil
}
