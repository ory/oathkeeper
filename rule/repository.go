// Copyright © 2022 Ory Corp

package rule

import (
	"context"

	"github.com/ory/oathkeeper/driver/configuration"
)

type Repository interface {
	List(ctx context.Context, limit, offset int) ([]Rule, error)
	Set(context.Context, []Rule) error
	Get(context.Context, string) (*Rule, error)
	Count(context.Context) (int, error)
	MatchingStrategy(context.Context) (configuration.MatchingStrategy, error)
	SetMatchingStrategy(context.Context, configuration.MatchingStrategy) error
}
