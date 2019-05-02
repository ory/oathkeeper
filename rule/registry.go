package rule

import "github.com/ory/oathkeeper/x"

type internalRegistry interface {
	Registry
	x.RegistryWriter
}

type Registry interface {
	RuleValidator() Validator
	RuleManager() Repository
}
