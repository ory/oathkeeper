package rsakey

import (
	"crypto/rsa"
	"net/http"

	"encoding/json"

	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/ory/hydra/sdk/go/hydra/swagger"
	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2"
)

type HydraManager struct {
	key *rsa.PrivateKey
	SDK hydra.SDK
	Set string
}

func (m *HydraManager) Refresh() error {
	_, response, err := m.SDK.GetJsonWebKey("private", m.Set)
	if err != nil {
		return errors.WithStack(err)
	} else if response.StatusCode == http.StatusNotFound {
		response.Body.Close()

		_, response, err = m.SDK.CreateJsonWebKeySet(m.Set, swagger.JsonWebKeySetGeneratorRequest{
			Alg: "RS256",
		})
		if err != nil {
			return errors.WithStack(err)
		} else if response.StatusCode != http.StatusCreated {
			return errors.Errorf("Expected status code %d but got %d", http.StatusOK, response.StatusCode)
		}
	} else if response.StatusCode != http.StatusOK {
		return errors.Errorf("Expected status code %d but got %d", http.StatusOK, response.StatusCode)
	}

	payload := response.Payload
	set := new(jose.JSONWebKeySet)

	if err := json.Unmarshal(payload, set); err != nil {
		return errors.WithStack(err)
	}

	if len(set.Key("private")) < 1 {
		return errors.New("Expected at least one private key but got none")
	}

	privateKey, ok := set.Key("private")[0].Key.(*rsa.PrivateKey)
	if !ok {
		return errors.Errorf("Type assertion to *rsa.PrivateKey failed, make sure you are actually sending a RSA private key")
	}

	m.key = privateKey

	return nil
}

func (m *HydraManager) PublicKey() (*rsa.PublicKey, error) {
	if m.key == nil {
		if err := m.Refresh(); err != nil {
			return nil, err
		}
	}
	return &m.key.PublicKey, nil
}

func (m *HydraManager) PrivateKey() (*rsa.PrivateKey, error) {
	if m.key == nil {
		if err := m.Refresh(); err != nil {
			return nil, err
		}
	}
	return m.key, nil
}

func (m *HydraManager) Algorithm() string {
	return "RS256"
}
