package authn

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"

	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/x"
)

type AuthenticatorForcedResponse struct {
	redirect string
}

func NewAuthenticatorForcedResponse(redirect string) *AuthenticatorForcedResponse {
	return &AuthenticatorForcedResponse{redirect: redirect}
}

func (a *AuthenticatorForcedResponse) Validate() error {
	return nil
}

func (a *AuthenticatorForcedResponse) GetID() string {
	if len(a.redirect) > 0 {
		return "redirect"
	} else {
		return "response"
	}
}

func (a *AuthenticatorForcedResponse) Authenticate(r *http.Request, config json.RawMessage, _ pipeline.Rule) (*AuthenticationSession, error) {
	if len(a.redirect) > 0 {
		rw := x.NewSimpleResponseWriter()
		http.Redirect(rw, r, a.redirect, http.StatusFound)
		*r = *r.WithContext(context.WithValue(r.Context(), pipeline.DirectorForcedResponse, &http.Response{
			StatusCode: rw.StatusCode,
			Body:       ioutil.NopCloser(new(bytes.Buffer)),
			Header:     rw.Header(),
		}))
		return nil, errors.WithStack(helper.ErrForceResponse)
	} else {
		*r = *r.WithContext(context.WithValue(r.Context(), pipeline.DirectorForcedResponse, &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString("response")),
			Header:     http.Header{"Content-Type": {"text/text"}},
		}))
		return nil, errors.WithStack(helper.ErrForceResponse)
	}
}
