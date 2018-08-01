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

package rule

import (
	"fmt"

	"github.com/asaskevich/govalidator"
	"github.com/ory/go-convenience/stringslice"
	"github.com/ory/oathkeeper/helper"
	"github.com/pkg/errors"
)

func ValidateRule(
	enabledAuthenticators []string, availableAuthenticators []string,
	enabledAuthorizers []string, availableAuthorizers []string,
	enabledCredentialsIssuers []string, availableCredentialsIssuers []string,
) func(r *Rule) error {
	methods := []string{"GET", "POST", "PUT", "HEAD", "DELETE", "PATCH", "OPTIONS", "TRACE", "CONNECT"}

	return func(r *Rule) error {
		// This is disabled because it doesn't support checking for regular expressions (obviously).
		// if !govalidator.IsURL(r.Match.URL) {
		// 	 return errors.WithStack(helper.ErrBadRequest.WithReason(fmt.Sprintf("Value \"%s\" from match.url field is not a valid url.", r.Match.URL)))
		// }

		for _, m := range r.Match.Methods {
			if !stringslice.Has(methods, m) {
				return errors.WithStack(helper.ErrBadRequest.WithReason(fmt.Sprintf("Value \"%s\" from match.methods is not a valid HTTP method, valid methods are: %v", m, methods)))
			}
		}

		if !govalidator.IsURL(r.Upstream.URL) {
			return errors.WithStack(helper.ErrBadRequest.WithReason(fmt.Sprintf("Value \"%s\" from upstream.url field is not a valid url.", r.Upstream.URL)))
		}

		if len(r.Authenticators) == 0 {
			return errors.WithStack(helper.ErrBadRequest.WithReason("At least one authenticator must be set."))
		}

		for _, a := range r.Authenticators {
			if !stringslice.Has(enabledAuthenticators, a.Handler) {
				if stringslice.Has(availableAuthenticators, a.Handler) {
					return errors.WithStack(helper.ErrBadRequest.WithReason(fmt.Sprintf("Authenticator \"%s\" is valid but has not enabled by the server's configuration, enabled authorizers are: %v", a.Handler, enabledAuthenticators)))
				}

				return errors.WithStack(helper.ErrBadRequest.WithReason(fmt.Sprintf("Authenticator \"%s\" is unknown, enabled authenticators are: %v", a.Handler, enabledAuthenticators)))
			}
		}

		if r.Authorizer.Handler == "" {
			return errors.WithStack(helper.ErrBadRequest.WithReason("Value authorizer.handler can not be empty."))
		}

		if !stringslice.Has(enabledAuthorizers, r.Authorizer.Handler) {
			if stringslice.Has(availableAuthorizers, r.Authorizer.Handler) {
				return errors.WithStack(helper.ErrBadRequest.WithReason(fmt.Sprintf("Authorizer \"%s\" is valid but has not enabled by the server's configuration, enabled authorizers are: %v", r.Authorizer.Handler, enabledAuthorizers)))
			}

			return errors.WithStack(helper.ErrBadRequest.WithReason(fmt.Sprintf("Authorizer \"%s\" is unknown, enabled authorizers are: %v", r.Authorizer.Handler, enabledAuthorizers)))
		}

		if r.CredentialsIssuer.Handler == "" {
			return errors.WithStack(helper.ErrBadRequest.WithReason("Value credentials_issuer.handler can not be empty."))
		}

		if !stringslice.Has(enabledCredentialsIssuers, r.CredentialsIssuer.Handler) {
			if stringslice.Has(availableCredentialsIssuers, r.CredentialsIssuer.Handler) {
				return errors.WithStack(helper.ErrBadRequest.WithReason(fmt.Sprintf("Credentials issuer \"%s\" is valid but has not enabled by the server's configuration, enabled credentials issuers are: %v", r.CredentialsIssuer.Handler, enabledCredentialsIssuers)))
			}

			return errors.WithStack(helper.ErrBadRequest.WithReason(fmt.Sprintf("Credentials issuer \"%s\" is unknown, enabled credentials issuers are: %v", r.CredentialsIssuer.Handler, enabledCredentialsIssuers)))
		}

		return nil
	}
}
