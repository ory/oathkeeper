package telemetry

import (
	"runtime"
	"time"

	"github.com/ory/hydra/pkg"
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
