// Copyright 2021 Ory GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package readiness

import (
	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/driver/health"
)

type (
	RuleReadinessChecker struct {
		hasFirstRuleLoad bool
	}

	RuleLoadedEvent struct{}
)

const ProbeName = "rule-first-load"

var ErrRuleNotYetLoaded = errors.New("rules have not been loaded yet")

func NewReadinessHealthChecker() *RuleReadinessChecker {
	return &RuleReadinessChecker{
		hasFirstRuleLoad: false,
	}
}

func (r *RuleReadinessChecker) ID() string {
	return ProbeName
}

func (r *RuleReadinessChecker) Validate() error {
	if !r.hasFirstRuleLoad {
		return errors.WithStack(ErrRuleNotYetLoaded)
	}
	return nil
}

func (r *RuleReadinessChecker) EventTypes() []health.ReadinessProbeEvent {
	return []health.ReadinessProbeEvent{&RuleLoadedEvent{}}
}

func (r *RuleReadinessChecker) EventsReceiver(event health.ReadinessProbeEvent) {
	switch event.(type) {
	case *RuleLoadedEvent:
		r.hasFirstRuleLoad = true
	}
}

func (r *RuleLoadedEvent) ReadinessProbeListenerID() string {
	return ProbeName
}
