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
