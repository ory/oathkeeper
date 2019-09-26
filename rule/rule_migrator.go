package rule

import (
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

	if semver.MustParseRange(">=0.19.0-alpha.0")(version) {
		return raw, nil
	}

	return nil, errors.WithStack(herodot.ErrBadRequest.WithReasonf("Unknown access rule version %s, unable to migrate.", version.String()))
}
