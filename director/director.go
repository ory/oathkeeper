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
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func NewDirector(target *url.URL, eval evaluator.Evaluator, logger logrus.FieldLogger, secret string) *Director {
	if logger == nil {
		logger = logrus.New()
	}
	if secret == "" {
		secret = uuid.New()
		logger.WithField("secret", secret).Infoln("No JWT secret was found, generated a random one.")
	}
	return &Director{
		TargetURL: target,
		Logger:    logger,
		Evaluator: eval,
		Secret:    secret,
	}
}

type Director struct {
	TargetURL *url.URL
	Logger    logrus.FieldLogger
	Evaluator evaluator.Evaluator
	Secret    string
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
			d.Logger.WithField("url", r.URL.String()).WithError(err).Print("Round trip failed.")
		}

		return res, err
	} else if _, ok := r.Context().Value(requestBypassedAuthorization).(string); ok {
		res, err := http.DefaultTransport.RoundTrip(r)
		if err != nil {
			d.Logger.WithField("url", r.URL.String()).WithError(err).Print("Round trip failed.")
		}

		return res, err
	}

	d.Logger.WithFields(map[string]interface{}{"user": "anonymous", "request_url": r.URL.String()}).Info("Unable to type assert context.")
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

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, access.ToClaims()).SignedString([]byte(d.Secret))
	if err != nil {
		d.Logger.
			WithError(errors.WithStack(err)).
			WithFields(map[string]interface{}{"user": access.User, "client_id": access.ClientID, "request_url": r.URL.String()}).
			Errorf("Unable to sign JSON Web Token.")
		*r = *r.WithContext(context.WithValue(r.Context(), requestDenied, &directorError{err: errors.WithStack(err), statusCode: http.StatusInternalServerError}))
		return
	}

	r.URL.Scheme = d.TargetURL.Scheme
	r.URL.Host = d.TargetURL.Host
	*r = *r.WithContext(context.WithValue(r.Context(), requestAllowed, token))
}
