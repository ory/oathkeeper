/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

package rsakey

import (
	"log"
	"net/http"
	"os"
	"testing"

	"time"

	"crypto/rsa"

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
		"local_rs256": &LocalRS256Manager{
			KeyStrength: 512,
		},
		"local_hs256": NewLocalHS256Manager([]byte("foobarbaz")),
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
			pub, err := m.PublicKey()
			require.NoError(t, err)
			require.NotNil(t, pub)

			switch pub.(type) {
			case *rsa.PublicKey:
				assert.Equal(t, "RS256", m.Algorithm())
			case []byte:
				assert.Equal(t, "HS256", m.Algorithm())
			}

			priv, err := m.PrivateKey()
			require.NoError(t, err)
			assert.NotNil(t, priv)
		})
	}
}

func connectToHydra(t *testing.T) *hydra.CodeGenSDK {
	if url := os.Getenv("TEST_HYDRA_URL"); url != "" {
		sdk, err := hydra.NewSDK(&hydra.Configuration{
			AdminURL: url,
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
		Tag:          "v1.0.0-beta.8",
		Cmd:          []string{"serve", "all", "--dangerous-force-http"},
		Env:          []string{"DATABASE_URL=memory"},
		ExposedPorts: []string{"4444/tcp", "4445/tcp"},
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	pool.MaxWait = time.Minute * 2

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
		AdminURL:  "http://localhost:" + resource.GetPort("4445/tcp") + "/",
		PublicURL: "http://localhost:" + resource.GetPort("4444/tcp") + "/",
	})
	require.NoError(t, err)
	return sdk
}
