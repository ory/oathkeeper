package rule

import (
	"net/url"

	"github.com/pkg/errors"
)

type Manager interface {
	ListRules() ([]Rule, error)
	CreateRule(*Rule) error
	GetRule(id string) (*Rule, error)
	DeleteRule(id string) error
	UpdateRule(*Rule) error
}

func NewManager(db string) (Manager, error) {
	if db == "memory" {
		return &MemoryManager{Rules: map[string]Rule{}}, nil
	} else if db == "" {
		return nil, errors.New("No database URL provided")
	}

	u, err := url.Parse(db)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	switch u.Scheme {
	case "postgres":
	case "mysql":
	}

	return nil, errors.Errorf("The provided database URL %s can not be handled", db)
}
