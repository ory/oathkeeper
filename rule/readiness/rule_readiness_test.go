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
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestReadinessHealthChecker(t *testing.T) {
	t.Run("rule readiness probe", func(t *testing.T) {
		ruleReadinessProbe := NewReadinessHealthChecker()
		ruleLoadedEvent := RuleLoadedEvent{}

		assert.Equal(t, ruleReadinessProbe.ID(), ProbeName)
		assert.Equal(t, ruleLoadedEvent.ReadinessProbeListenerID(), ProbeName)

		assert.True(t, errors.Is(ruleReadinessProbe.Validate(), ErrRuleNotYetLoaded))

		evtTypes := ruleReadinessProbe.EventTypes()
		assert.Len(t, evtTypes, 1)
		_, ok := evtTypes[0].(*RuleLoadedEvent)
		assert.True(t, ok, "actual type %T", evtTypes[0])

		// Dispatch fake event
		ruleReadinessProbe.EventsReceiver(&RuleLoadedEvent{})

		assert.NoError(t, ruleReadinessProbe.Validate())
	})
}
