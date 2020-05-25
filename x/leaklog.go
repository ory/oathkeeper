package x

import (
	"github.com/sirupsen/logrus"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/pipeline"
)

func RedactInProd(d configuration.Provider, value interface{}) interface{} {
	if d.IsInsecureDevMode() {
		return value
	}
	return "This value has been redacted to prevent leak of sensitive information to logs. Switch to ORY Kratos Development Mode using --dev to view the original value."
}

type LoggingProvider interface {
	Logger() logrus.FieldLogger
}

func LogAccessRuleContext(d LoggingProvider, c configuration.Provider, r pipeline.Rule, context interface{}) logrus.FieldLogger {
	return d.Logger().WithFields(logrus.Fields{
		"context": RedactInProd(c, context),
		"rule_id": r.GetID(),
		"handler": "authz/remote"})
}