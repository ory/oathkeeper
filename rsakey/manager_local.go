package rsakey

import (
	"crypto/rand"
	"crypto/rsa"

	"github.com/pkg/errors"
)

type LocalManager struct {
	key         *rsa.PrivateKey
	KeyStrength int
}

func (m *LocalManager) Refresh() error {
	if m.KeyStrength == 0 {
		m.KeyStrength = 4096
	}

	key, err := rsa.GenerateKey(rand.Reader, m.KeyStrength)
	if err != nil {
		return errors.WithStack(err)
	}

	m.key = key
	return nil
}

func (m *LocalManager) PublicKey() (*rsa.PublicKey, error) {
	if m.key == nil {
		if err := m.Refresh(); err != nil {
			return nil, err
		}
	}
	return &m.key.PublicKey, nil
}

func (m *LocalManager) PrivateKey() (*rsa.PrivateKey, error) {
	if m.key == nil {
		if err := m.Refresh(); err != nil {
			return nil, err
		}
	}
	return m.key, nil
}

func (m *LocalManager) Algorithm() string {
	return "RS256"
}
