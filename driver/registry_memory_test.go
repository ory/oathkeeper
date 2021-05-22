package driver

import (
	"testing"

	"github.com/ory/x/logrusx"
	"github.com/stretchr/testify/assert"

	"github.com/ory/oathkeeper/driver/configuration"
)

func TestRegistryMemoryAvailablePipelineAuthorizers(t *testing.T) {
	r := NewRegistryMemory().WithConfig(configuration.NewViperProvider(logrusx.New("test", "0.0")))
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
			r := NewRegistryMemory().WithConfig(configuration.NewViperProvider(logrusx.New("test", "0.0")))
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
