// Copyright 2021 Ory GmbH
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

package driver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistryMemoryAvailablePipelineAuthorizers(t *testing.T) {
	r := NewRegistryMemory()
	got := r.AvailablePipelineAuthorizers()
	assert.ElementsMatch(t, got, []string{"allow", "deny", "keto_engine_acp_ory", "remote", "remote_json"})
}

func TestRegistryMemoryPipelineAuthorizer(t *testing.T) {
	tests := []struct {
		id      string
		wantErr bool
	}{
		{id: "allow"},
		{id: "deny"},
		{id: "keto_engine_acp_ory"},
		{id: "remote"},
		{id: "remote_json"},
		{id: "unregistered", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			r := NewRegistryMemory()
			a, err := r.PipelineAuthorizer(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("PipelineAuthorizer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if a != nil && a.GetID() != tt.id {
				t.Errorf("PipelineAuthorizer() got = %v, want %v", a.GetID(), tt.id)
			}
		})
	}
}
