package authn

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hashicorp/golang-lru"
	"github.com/ory/herodot"
	"github.com/ory/hive-cloud/hive/auth"
	"github.com/ory/x/httpx"
	"github.com/ory/x/urlx"
	"github.com/pkg/errors"

	"github.com/ory/hive-cloud/hive/session"

	"github.com/ory/oathkeeper/driver/configuration"
	"github.com/ory/oathkeeper/helper"
	"github.com/ory/oathkeeper/pipeline"
	"github.com/ory/oathkeeper/x"
)

type AuthenticatorHive struct {
	c configuration.Provider

	sessions *lru.ARCCache
	client   *http.Client
}

type authenticatorHiveConfiguration struct {
	OnUnauthorized string `json:"on_unauthorized"`
	CookieName     string `json:"cookie_name"`
}

func (c *authenticatorHiveConfiguration) getCookieName() string {
	if len(c.CookieName) == 0 {
		return session.DefaultSessionCookieName
	}
	return c.CookieName
}

func NewAuthenticatorHive(c configuration.Provider) *AuthenticatorHive {
	arc, _ := lru.NewARC(1024) // Error can be ignored because only thrown when negative size
	return &AuthenticatorHive{
		sessions: arc,
		c:        c,
		client:   httpx.NewResilientClientLatencyToleranceMedium(nil),
	}
}

func (a *AuthenticatorHive) GetID() string {
	return "hive"
}

func (a *AuthenticatorHive) Validate() error {
	if !a.c.AuthenticatorHiveIsEnabled() {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf(`Authenticator "%s" is disabled per configuration.`, a.GetID()))
	}

	if a.c.AuthenticatorHiveAdminURL() == nil {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf(`Configuration for authenticator "%s" did not specify any values for configuration key "%s" and is thus disabled.`, a.GetID(), configuration.ViperKeyAuthenticatorHiveAdminURL))
	}

	if a.c.AuthenticatorHivePublicURL() == nil {
		return errors.WithStack(ErrAuthenticatorNotEnabled.WithReasonf(`Configuration for authenticator "%s" did not specify any values for configuration key "%s" and is thus disabled.`, a.GetID(), configuration.ViperKeyAuthenticatorHivePublicURL))
	}

	return nil
}

func (a *AuthenticatorHive) Authenticate(r *http.Request, config json.RawMessage, rl pipeline.Rule) (*AuthenticationSession, error) {
	var cf authenticatorHiveConfiguration
	if len(config) == 0 {
		config = []byte("{}")
	}

	d := json.NewDecoder(bytes.NewBuffer(config))
	d.DisallowUnknownFields()
	if err := d.Decode(&cf); err != nil {
		return nil, errors.WithStack(err)
	}

	cookie, err := r.Cookie(cf.getCookieName())
	if err == http.ErrNoCookie {
		return nil, a.handleUnauthorized(r, &cf)
	} else if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithDebug(err.Error()))
	}

	s, err := a.FindSession(cookie)
	if errors.Cause(err) == helper.ErrUnauthorized {
		return nil, a.handleUnauthorized(r, &cf)
	} else if err != nil {
		return nil, err
	}

	// Todo check for AuthenticationFactor
	// Todo check if we want to reauthenticate because the auth time is too short
	// Todo update session (?) , check LastUpdatedAt

	var b bytes.Buffer
	extra := map[string]interface{}{}
	if err := json.NewEncoder(&b).Encode(s); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithDebug(err.Error()))
	}

	if err := json.NewDecoder(&b).Decode(&extra); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithDebug(err.Error()))
	}

	return &AuthenticationSession{
		Subject: s.Identity.ID,
		Extra:   extra,
	}, nil
}

func (a *AuthenticatorHive) FindSession(c *http.Cookie) (*session.Session, error) {
	s, ok := a.sessions.Get(c.Value)
	if ok {
		if ss, valid := s.(session.Session); !valid {
			return nil, errors.WithStack(herodot.ErrInternalServerError.WithErrorf("Expected type of session to be *AuthenticatorHiveSession but got: %T", s))
		} else {
			return &ss, nil
		}
	}

	req, err := http.NewRequest("GET", urlx.AppendPaths(a.c.AuthenticatorHiveAdminURL(), session.SessionMePath).String(), nil)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithDebug(err.Error()))
	}
	req.AddCookie(c)

	res, err := a.client.Do(req)
	if err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithDebug(err.Error()))
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return nil, errors.WithStack(helper.ErrUnauthorized)
	} else if res.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithReasonf("Expected status code 200 from session url but got: %d", res.StatusCode).WithDebugf("Response body was: %s", body))
	}

	var ss session.Session
	d := json.NewDecoder(res.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&ss); err != nil {
		return nil, errors.WithStack(herodot.ErrInternalServerError.WithDebug(err.Error()))
	}

	a.sessions.Add(c.Value, ss)

	return &ss, nil
}

func (a *AuthenticatorHive) handleUnauthorized(r *http.Request, cf *authenticatorHiveConfiguration) error {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(&struct {
		RedirectTo string `json:"redirect_to"`
	}{
		RedirectTo: urlx.AppendPaths(a.c.AuthenticatorHivePublicURL(), auth.BrowserSignInPath).String(),
	}); err != nil {
		return errors.WithStack(herodot.ErrInternalServerError.WithDebug(err.Error()))
	}

	switch cf.OnUnauthorized {
	case "reject":
		*r = *r.WithContext(context.WithValue(r.Context(), pipeline.DirectorForcedResponse, &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       ioutil.NopCloser(&b),
			Header:     http.Header{"Content-Type": {"application/json"}},
		}))
		return errors.WithStack(helper.ErrForceResponse)
	case "redirect":
		rw := x.NewSimpleResponseWriter()
		http.Redirect(rw, r, urlx.AppendPaths(a.c.AuthenticatorHivePublicURL(), auth.BrowserSignInPath).String(), http.StatusFound)
		*r = *r.WithContext(context.WithValue(r.Context(), pipeline.DirectorForcedResponse, &http.Response{
			StatusCode: rw.StatusCode,
			Body:       ioutil.NopCloser(new(bytes.Buffer)),
			Header:     rw.Header(),
		}))
		return errors.WithStack(helper.ErrForceResponse)
	case "dynamic":
		fallthrough
	default:
		rw := x.NewSimpleResponseWriter()
		if isAPIRequest(r) {
			*r = *r.WithContext(context.WithValue(r.Context(), pipeline.DirectorForcedResponse, &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       ioutil.NopCloser(&b),
				Header:     http.Header{"Content-Type": {"application/json"}},
			}))

			return errors.WithStack(helper.ErrForceResponse)
		}

		http.Redirect(rw, r, urlx.AppendPaths(a.c.AuthenticatorHivePublicURL(), auth.BrowserSignInPath).String(), http.StatusFound)
		*r = *r.WithContext(context.WithValue(r.Context(), pipeline.DirectorForcedResponse, &http.Response{
			StatusCode: rw.StatusCode,
			Body:       ioutil.NopCloser(new(bytes.Buffer)),
			Header:     rw.Header(),
		}))
		return errors.WithStack(helper.ErrForceResponse)
	}
}

func isAPIRequest(r *http.Request) bool {
	if r.Header.Get("Content-Type") == "application/json" {
		return true
	} else if r.Header.Get("Content-Type") == "application/xml" {
		return true
	} else if len(r.Header.Get("Origin")) > 0 {
		return true
	} else if strings.ToLower(r.Header.Get("HTTP_X_REQUESTED_WITH")) == "xmlhttprequest" {
		return true
	}

	// Todo: Maybe do some content negotiation with Accept header?
	// application/json
	// text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8
	// text/css
	// js -> */*

	return false
}
