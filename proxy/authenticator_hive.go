package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"

	"github.com/ory/hive-cloud/hive/session"
	"github.com/ory/oathkeeper/helper"

	"github.com/ory/oathkeeper/rule"
)

type AuthenticatorHive struct {
	sessions       *lru.ARCCache
	client         *http.Client
	sessionURL     string
	sessionInitURL string
}

type authenticatorHiveConfiguration struct {
	OnUnauthorized string `json:"on_unauthorized"`
	CookieName     string `json:"cookie_name"`
}

type authenticatorHiveConfigurationOnUnauthorized int

const (
	authenticatorHiveConfigurationOnUnauthorizedReject authenticatorHiveConfigurationOnUnauthorized = iota + 1
	authenticatorHiveConfigurationOnUnauthorizedRedirect
	authenticatorHiveConfigurationOnUnauthorizedDynamic
)

func (c *authenticatorHiveConfiguration) onUnauthorized() (authenticatorHiveConfigurationOnUnauthorized, error) {
	if c.OnUnauthorized == "" {
		return authenticatorHiveConfigurationOnUnauthorizedDynamic, nil
	}

	modes := map[string]authenticatorHiveConfigurationOnUnauthorized{
		"reject":   authenticatorHiveConfigurationOnUnauthorizedReject,
		"redirect": authenticatorHiveConfigurationOnUnauthorizedRedirect,
		"dynamic":  authenticatorHiveConfigurationOnUnauthorizedDynamic,
	}

	if mode, ok := modes[strings.ToLower(c.OnUnauthorized)]; !ok {
		return 0, helper.ErrRuleMisconfiguration.WithReasonf(`Expected field on_authorized to contain one of "redirect", "dynamic", "reject", or "" but got: %s`, c.OnUnauthorized)
	} else {
		return mode, nil
	}
}

type authenticatorHiveUnauthorizedResponse struct {
	RedirectTo string `json:"redirect_to"`
}

func (c *authenticatorHiveConfiguration) getCookieName() string {
	if len(c.CookieName) == 0 {
		return session.DefaultSessionCookieName
	}
	return c.CookieName
}

func NewAuthenticatorHive(c *http.Client, sessionURL, sessionInitURL string) *AuthenticatorHive {
	arc, _ := lru.NewARC(1024) // Error can be ignored because only thrown when negative size
	return &AuthenticatorHive{
		sessions:       arc,
		client:         c,
		sessionURL:     sessionURL,
		sessionInitURL: sessionInitURL,
	}
}

func (a *AuthenticatorHive) GetID() string {
	return "hive"
}

func (a *AuthenticatorHive) findSession(c *http.Cookie) (*session.Session, error) {
	s, ok := a.sessions.Get(c.Value)
	if ok {
		if ss, valid := s.(session.Session); !valid {
			return nil, errors.WithStack(helper.ErrServerError.WithErrorf("Expected type of session to be *AuthenticatorHiveSession but got: %T", s))
		} else {
			return &ss, nil
		}
	}

	req, err := http.NewRequest("GET", a.sessionURL, nil)
	if err != nil {
		return nil, errors.WithStack(helper.ErrServerError.WithDebug(err.Error()))
	}
	req.AddCookie(c)

	res, err := a.client.Do(req)
	if err != nil {
		return nil, errors.WithStack(helper.ErrServerError.WithDebug(err.Error()))
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return nil, errors.WithStack(helper.ErrUnauthorized)
	} else if res.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)
		return nil, errors.WithStack(helper.ErrServerError.WithReasonf("Expected status code 200 from session url but got: %d", res.StatusCode).WithDebugf("Response body was: %s", body))
	}

	var ss session.Session
	d := json.NewDecoder(res.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&ss); err != nil {
		return nil, errors.WithStack(helper.ErrServerError.WithDebug(err.Error()))
	}

	a.sessions.Add(c.Value, ss)

	return &ss, nil
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

func (a *AuthenticatorHive) handleUnauthorized(r *http.Request, cf *authenticatorHiveConfiguration) error {
	mode, err := cf.onUnauthorized()
	if err != nil {
		return err
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(&authenticatorHiveUnauthorizedResponse{
		RedirectTo: a.sessionInitURL,
	}); err != nil {
		return errors.WithStack(helper.ErrServerError.WithDebug(err.Error()))
	}

	switch mode {
	case authenticatorHiveConfigurationOnUnauthorizedReject:
		*r = *r.WithContext(context.WithValue(r.Context(), directorForcedResponse, &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       ioutil.NopCloser(&b),
			Header:     http.Header{"Content-Type": {"application/json"}},
		}))

		return errors.WithStack(helper.ErrForceResponse)
	case authenticatorHiveConfigurationOnUnauthorizedRedirect:
		rw := NewSimpleResponseWriter()
		http.Redirect(rw, r, a.sessionInitURL, http.StatusFound)
		*r = *r.WithContext(context.WithValue(r.Context(), directorForcedResponse, &http.Response{
			StatusCode: rw.code,
			Body:       ioutil.NopCloser(new(bytes.Buffer)),
			Header:     rw.header,
		}))
		return errors.WithStack(helper.ErrForceResponse)
	case authenticatorHiveConfigurationOnUnauthorizedDynamic:
		rw := NewSimpleResponseWriter()
		if isAPIRequest(r) {
			*r = *r.WithContext(context.WithValue(r.Context(), directorForcedResponse, &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       ioutil.NopCloser(&b),
				Header:     http.Header{"Content-Type": {"application/json"}},
			}))

			return errors.WithStack(helper.ErrForceResponse)
		}

		http.Redirect(rw, r, a.sessionInitURL, http.StatusFound)
		*r = *r.WithContext(context.WithValue(r.Context(), directorForcedResponse, &http.Response{
			StatusCode: rw.code,
			Body:       ioutil.NopCloser(new(bytes.Buffer)),
			Header:     rw.header,
		}))
		return errors.WithStack(helper.ErrForceResponse)
	default:
		panic(fmt.Sprintf("An unknown mode %s was used which should not happen as the modes have been validated", cf.OnUnauthorized))
	}
}

func (a *AuthenticatorHive) Authenticate(r *http.Request, config json.RawMessage, rl *rule.Rule) (*AuthenticationSession, error) {
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
		return nil, errors.WithStack(helper.ErrServerError.WithDebug(err.Error()))
	}

	s, err := a.findSession(cookie)
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
		return nil, errors.WithStack(helper.ErrServerError.WithDebug(err.Error()))
	}

	if err := json.NewDecoder(&b).Decode(&extra); err != nil {
		return nil, errors.WithStack(helper.ErrServerError.WithDebug(err.Error()))
	}

	return &AuthenticationSession{
		Subject: s.Identity.ID,
		Extra:   extra,
	}, nil
}
