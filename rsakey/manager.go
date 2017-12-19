package rsakey

import "crypto/rsa"

type Manager interface {
	Refresh() error
	PrivateKey() (*rsa.PrivateKey, error)
	PublicKey() (*rsa.PublicKey, error)
	PublicKeyID() string
	Algorithm() string
}
