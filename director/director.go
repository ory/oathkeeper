// Copyright © 2017-2018 Aeneas Rekkas <aeneas+oss@aeneas.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package director

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/dgrijalva/jwt-go"
	"github.com/ory/oathkeeper/evaluator"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/rsakey"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func NewDirector(target *url.URL, eval evaluator.Evaluator, logger logrus.FieldLogger, keyManager rsakey.Manager) *Director {
	if logger == nil {
		logger = logrus.New()
	}
	return &Director{
		TargetURL:  target,
		Logger:     logger,
		Evaluator:  eval,
		KeyManager: keyManager,
	}
}

type Director struct {
	TargetURL  *url.URL
	Logger     logrus.FieldLogger
	Evaluator  evaluator.Evaluator
	KeyManager rsakey.Manager
}

type key int

const requestAllowed key = 0
const requestDenied key = 1
const requestBypassedAuthorization key = 2

type directorError struct {
	err        error
	statusCode int
}

func (d *Director) RoundTrip(r *http.Request) (*http.Response, error) {
	if err, ok := r.Context().Value(requestDenied).(*directorError); ok && err != nil {
		return &http.Response{
			StatusCode: err.statusCode,
			Body:       ioutil.NopCloser(bytes.NewBufferString(err.err.Error())),
		}, nil
	} else if token, ok := r.Context().Value(requestAllowed).(string); ok {
		r.Header.Set("Authorization", "bearer "+token)
		res, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			d.Logger.WithField("url", r.URL.String()).WithError(err).Print("Round trip failed")
		}

		return res, err
	} else if _, ok := r.Context().Value(requestBypassedAuthorization).(string); ok {
		res, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			d.Logger.WithField("url", r.URL.String()).WithError(err).Print("Round trip failed")
		}

		return res, err
	}

	d.Logger.WithFields(map[string]interface{}{"user": "anonymous", "request_url": r.URL.String()}).Info("Unable to type assert context")
	return &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       ioutil.NopCloser(bytes.NewBufferString(http.StatusText(http.StatusInternalServerError))),
	}, nil
}

func (d *Director) Director(r *http.Request) {
	access, err := d.Evaluator.EvaluateAccessRequest(r)
	if err != nil {
		switch errors.Cause(err) {
		case helper.ErrForbidden:
			*r = *r.WithContext(context.WithValue(r.Context(), requestDenied, &directorError{err: err, statusCode: http.StatusForbidden}))
		case helper.ErrMissingBearerToken:
			*r = *r.WithContext(context.WithValue(r.Context(), requestDenied, &directorError{err: err, statusCode: http.StatusUnauthorized}))
		case helper.ErrUnauthorized:
			*r = *r.WithContext(context.WithValue(r.Context(), requestDenied, &directorError{err: err, statusCode: http.StatusUnauthorized}))
		case helper.ErrMatchesNoRule:
			*r = *r.WithContext(context.WithValue(r.Context(), requestDenied, &directorError{err: err, statusCode: http.StatusNotFound}))
		case helper.ErrMatchesMoreThanOneRule:
			*r = *r.WithContext(context.WithValue(r.Context(), requestDenied, &directorError{err: err, statusCode: http.StatusInternalServerError}))
		default:
			*r = *r.WithContext(context.WithValue(r.Context(), requestDenied, &directorError{err: err, statusCode: http.StatusInternalServerError}))
		}

		return
	}

	if access.Disabled {
		r.URL.Scheme = d.TargetURL.Scheme
		r.URL.Host = d.TargetURL.Host
		*r = *r.WithContext(context.WithValue(r.Context(), requestBypassedAuthorization, ""))
		return
	}

	privateKey, err := d.KeyManager.PrivateKey()
	if err != nil {
		d.Logger.
			WithError(errors.WithStack(err)).
			WithFields(map[string]interface{}{"user": access.User, "client_id": access.ClientID, "request_url": r.URL.String()}).
			Errorf("Unable to fetch private key for signing JSON Web Token")
		*r = *r.WithContext(context.WithValue(r.Context(), requestDenied, &directorError{err: errors.WithStack(err), statusCode: http.StatusInternalServerError}))
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, access.ToClaims())
	token.Header["kid"] = d.KeyManager.PublicKeyID()

	signed, err := token.SignedString(privateKey)
	if err != nil {
		d.Logger.
			WithError(errors.WithStack(err)).
			WithFields(map[string]interface{}{"user": access.User, "client_id": access.ClientID, "request_url": r.URL.String()}).
			Errorf("Unable to sign JSON Web Token")
		*r = *r.WithContext(context.WithValue(r.Context(), requestDenied, &directorError{err: errors.WithStack(err), statusCode: http.StatusInternalServerError}))
		return
	}

	r.URL.Scheme = d.TargetURL.Scheme
	r.URL.Host = d.TargetURL.Host
	*r = *r.WithContext(context.WithValue(r.Context(), requestAllowed, signed))
}
