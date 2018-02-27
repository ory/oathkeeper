/*
 * Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * @author       Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @copyright  2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
 * @license  	   Apache-2.0
 */

package telemetry

import (
	"runtime"
	"time"

	"strings"

	"github.com/ory/oathkeeper/pkg"
	"github.com/segmentio/analytics-go"
	"github.com/sirupsen/logrus"
)

type Manager struct {
	Segment      *analytics.Client
	Logger       logrus.FieldLogger
	ID           string
	InstanceID   string
	BuildVersion string
	BuildHash    string
	BuildTime    string
	Middleware   *Middleware
}

func (m *Manager) Identify() {
	if strings.Contains(m.ID, "localhost") {
		return
	}

	if err := pkg.Retry(m.Logger, time.Minute*2, time.Minute*15, func() error {
		return m.Segment.Identify(&analytics.Identify{
			UserId: m.ID,
			Traits: map[string]interface{}{
				"goarch":         runtime.GOARCH,
				"goos":           runtime.GOOS,
				"numCpu":         runtime.NumCPU(),
				"runtimeVersion": runtime.Version(),
				"version":        m.BuildVersion,
				"hash":           m.BuildHash,
				"buildTime":      m.BuildTime,
				"instance":       m.InstanceID,
			},
		})
	}); err != nil {
		m.Logger.WithError(err).Debug("Could not commit environment information")
	}
}

func (m *Manager) Submit() {
	if strings.Contains(m.ID, "localhost") {
		return
	}

	for {
		if err := m.Segment.Track(&analytics.Track{
			Event:  "telemetry",
			UserId: m.ID,
			Properties: map[string]interface{}{
				"requests": m.Middleware.Requests,
				"instance": m.InstanceID,
			},
		}); err != nil {
			m.Logger.WithError(err).Debug("Could not commit data")
		}
		time.Sleep(time.Hour)
	}
}
