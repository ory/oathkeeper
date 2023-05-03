// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"context"
	"net/http"

	"github.com/ory/oathkeeper/driver/configuration"
)

type Repository interface {
	List(ctx context.Context, limit, offset int) ([]Rule, error)
	Set(context.Context, []Rule) error
	Get(context.Context, string) (*Rule, error)
	Count(context.Context) (int, error)
	MatchingStrategy(context.Context) (configuration.MatchingStrategy, error)
	SetMatchingStrategy(context.Context, configuration.MatchingStrategy) error
	ReadyChecker(*http.Request) error
}
