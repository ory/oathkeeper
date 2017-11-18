package rsakey

import (
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/ory/dockertest"
	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var resources []*dockertest.Resource
var pool *dockertest.Pool

func killAll() {
	for _, resource := range resources {
		if err := pool.Purge(resource); err != nil {
			log.Printf("Got an error while trying to purge resource: %s", err)
		}
	}
	resources = []*dockertest.Resource{}
}

func TestMain(m *testing.M) {
	code := m.Run()
	killAll()
	os.Exit(code)
}

func TestManager(t *testing.T) {
	managers := map[string]Manager{
		"local": &LocalManager{
			KeyStrength: 512,
		},
	}

	if !testing.Short() {
		sdk := connectToHydra(t)
		managers["hydra"] = &HydraManager{
			SDK: sdk,
			Set: "test-key",
		}
	}

	for k, m := range managers {
		t.Run("case="+k, func(t *testing.T) {
			require.NoError(t, m.Refresh())
			assert.EqualValues(t, "RS256", m.Algorithm())

			pub, err := m.PublicKey()
			require.NoError(t, err)
			assert.NotNil(t, pub)

			priv, err := m.PrivateKey()
			require.NoError(t, err)
			assert.NotNil(t, priv)
		})
	}
}

func connectToHydra(t *testing.T) *hydra.CodeGenSDK {
	scopes := []string{"hydra.keys.*"}
	if url := os.Getenv("TEST_HYDRA_URL"); url != "" {
		sdk, err := hydra.NewSDK(&hydra.Configuration{
			EndpointURL:  url,
			ClientID:     os.Getenv("TEST_HYDRA_CLIENT_ID"),
			ClientSecret: os.Getenv("TEST_HYDRA_CLIENT_SECRET"),
			Scopes:       scopes,
		})
		require.NoError(t, err)
		return sdk
	}

	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "oryd/hydra",
		Tag:          "v0.10.0-alpha.18",
		Cmd:          []string{"host", "--dangerous-force-http"},
		Env:          []string{"DATABASE_URL=memory", "FORCE_ROOT_CLIENT_CREDENTIALS=root:secret"},
		ExposedPorts: []string{"4444/tcp"},
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err = pool.Retry(func() error {
		var err error
		u := "http://localhost:" + resource.GetPort("4444/tcp") + "/health/status"
		t.Logf("Trying to connect to ORY Hydra at %s", u)
		response, err := http.Get(u)
		if err != nil {
			t.Logf("Unable to connect to ORY Hydra at %s because: %s", u, err)
			return err
		} else if response.StatusCode != http.StatusOK {
			t.Logf("Unable to connect to ORY Hydra at %s because status code %d was received", u, err)
			return errors.Errorf("Expected status code 200 but got %d while connecting to %s", response.StatusCode, u)
		}

		return nil
	}); err != nil {
		pool.Purge(resource)
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resources = append(resources, resource)
	sdk, err := hydra.NewSDK(&hydra.Configuration{
		EndpointURL:  "http://localhost:" + resource.GetPort("4444/tcp") + "/",
		ClientID:     "root",
		ClientSecret: "secret",
		Scopes:       scopes,
	})
	require.NoError(t, err)
	return sdk
}
