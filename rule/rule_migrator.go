// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package rule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/ory/herodot"
	"github.com/ory/x/stringsx"

	"github.com/ory/oathkeeper/x"
)

func migrateRuleJSON(raw []byte) ([]byte, error) {
	rv := strings.TrimPrefix(
		stringsx.Coalesce(
			gjson.GetBytes(raw, "version").String(),
			x.Version,
			x.UnknownVersion,
		),
		"v",
	)

	if rv == x.UnknownVersion {
		return raw, nil
	}

	version, err := semver.Make(rv)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	raw, err = sjson.SetBytes(raw, "version", strings.Split(x.Version, "+")[0])
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if semver.MustParseRange("<=0.32.0-beta.1")(version) {
		// Applies the following patch:
		//
		// - number_of_retries (int) => give_up_after (duration, string, := number_of_retries*delay_in_milliseconds + "ms")
		// - delay_in_milliseconds (int) => max_delay (duration, string, := delay_in_milliseconds + "ms")
		if mutators := gjson.GetBytes(raw, `mutators`); mutators.Exists() {
			for key, value := range mutators.Array() {
				if value.Get("handler").String() != "hydrator" {
					continue
				}

				rj := gjson.GetBytes(raw, fmt.Sprintf(`mutators.%d.config.retry.number_of_retries`, key))
				dj := gjson.GetBytes(raw, fmt.Sprintf(`mutators.%d.config.retry.delay_in_milliseconds`, key))

				var delay = int64(100)
				var retries = int64(3) //nolint:ineffassign // legacy migration keeps variable for clarity
				var err error
				if dj.Exists() {
					delay = dj.Int()
					if raw, err = sjson.SetBytes(raw, fmt.Sprintf(`mutators.%d.config.retry.max_delay`, key), fmt.Sprintf("%dms", delay)); err != nil {
						return nil, errors.WithStack(err)
					}

					if raw, err = sjson.DeleteBytes(raw, fmt.Sprintf(`mutators.%d.config.retry.delay_in_milliseconds`, key)); err != nil {
						return nil, errors.WithStack(err)
					}
				}

				if rj.Exists() {
					retries = rj.Int()
					if raw, err = sjson.SetBytes(raw, fmt.Sprintf(`mutators.%d.config.retry.give_up_after`, key), fmt.Sprintf("%dms", retries*delay)); err != nil {
						return nil, errors.WithStack(err)
					}

					if raw, err = sjson.DeleteBytes(raw, fmt.Sprintf(`mutators.%d.config.retry.number_of_retries`, key)); err != nil {
						return nil, errors.WithStack(err)
					}
				}
			}
		}

		version, _ = semver.Make("0.32.0-beta.1")
	}

	if semver.MustParseRange("<=0.37.0")(version) {
		// Applies the following patch:
		//
		// - in the keto_engine_acp_ory authorizer we have to use go/template instaead of replacement syntax ('$x')
		// - "required_action": "my:action:$1" => "required_action": "my:action:{{ printIndex .MatchContext.RegexpCaptureGroups 0}}"
		if authorizer := gjson.GetBytes(raw, `authorizer`); authorizer.Exists() {
			if authorizer.Get("handler").String() == "keto_engine_acp_ory" {

				aj := gjson.GetBytes(raw, `authorizer.config.required_action`)
				rj := gjson.GetBytes(raw, `authorizer.config.required_resource`)

				re := regexp.MustCompile(`\$([0-9]+)`)
				var err error
				if aj.Exists() {
					result := re.ReplaceAllString(aj.Str, "{{ printIndex .MatchContext.RegexpCaptureGroups (sub $1 1 | int)}}")
					if raw, err = sjson.SetBytes(raw, `authorizer.config.required_action`, result); err != nil {
						return nil, errors.WithStack(err)
					}
				}

				if rj.Exists() {
					result := re.ReplaceAllString(rj.Str, "{{ printIndex .MatchContext.RegexpCaptureGroups (sub $1 1 | int)}}")
					if raw, err = sjson.SetBytes(raw, `authorizer.config.required_resource`, result); err != nil {
						return nil, errors.WithStack(err)
					}
				}
			}
		}

		version, _ = semver.Make("0.37.0")
	}

	if semver.MustParseRange(">=0.37.0")(version) {
		return raw, nil
	}

	return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unknown access rule version %s, unable to migrate.", version.String()))
}
