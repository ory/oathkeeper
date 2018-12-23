package proxy

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"github.com/gorilla/sessions"
	"github.com/ory/oathkeeper/helper"
	"github.com/pkg/errors"
	"net/http"
	"time"

	"github.com/ory/oathkeeper/rule"
)

type AuthenticatorHive struct {
	SessionStore sessions.Store
}

type AuthenticatorHiveSession struct {
	// SID is the session id.
	SID string

	// LastUpdatedAt indicates when the user metadata was updated last.
	LastUpdatedAt string

	// AuthenticatedAt says when the session was authenticated.
	AuthenticatedAt time.Time

	// ExpiresAt indicates when the session should expire. If this is time.Zero()
	// the session never expires.
	ExpiresAt time.Time

	// Subject is the identity.
	Subject string

	// AuthenticationFactor
	AuthenticationFactor []string

	// SessionMetadata
	SessionMetadata map[string]interface{}
}

func init() {
	gob.Register(new(AuthenticatorHiveSession))
}

type AuthenticatorHiveConfiguration struct {
	RedirectOnUnauthorized bool   `json:"redirect_on_unauthorized"`
	CookieName             string `json:"cookie_name"`
	CookieDomain           string `json:"cookie_domain"`
}

func (c *AuthenticatorHiveConfiguration) getCookieName() string {
	if len(c.CookieName) == 0 {
		return "hive_session"
	}
	return c.CookieName
}

func NewAuthenticatorHive(secret []byte) *AuthenticatorHive {
	return &AuthenticatorHive{
		SessionStore: sessions.NewCookieStore(secret),
	}
}

func (a *AuthenticatorHive) GetID() string {
	return "hive"
}

func (a *AuthenticatorHive) Authenticate(r *http.Request, config json.RawMessage, rl *rule.Rule) (*AuthenticationSession, error) {
	var cf AuthenticatorHiveConfiguration
	if len(config) == 0 {
		config = []byte("{}")
	}

	d := json.NewDecoder(bytes.NewBuffer(config))
	d.DisallowUnknownFields()
	if err := d.Decode(&cf); err != nil {
		return nil, errors.WithStack(err)
	}

	session, err := a.SessionStore.Get(r, cf.getCookieName())
	if err != nil {
		// The error can be ignored here, it indicates that no session exists yet, which is fine.
	}

	internal, ok := session.Values["s"].(*AuthenticatorHiveSession)
	if !ok || internal == nil {
		return nil, errors.WithStack(helper.ErrUnauthorized.WithReason("No session cookie found in request."))
	}

	return &AuthenticationSession{
		Subject: internal.Subject,
		Extra:   internal.SessionMetadata,
	}, nil
}
