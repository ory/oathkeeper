package director

import (
	"testing"
	"net/url"
	"github.com/golang/mock/gomock"
	"net/http"
	"github.com/stretchr/testify/assert"
	"fmt"
	"github.com/ory-am/hydra/firewall"
	"errors"
	"github.com/ory-am/hydra/oauth2"
)

func roundTrip(r *http.Request) (error) {
	err, _ := r.Context().Value(wasDenied).(error)
	return err
}

func TestDirector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	u, _ := url.Parse("http://localhost:1234")
	fw := NewMockFirewall(ctrl)
	it := NewMockIntrospector(ctrl)
	d := &Director{
		TargetURL: u,
		Firewall: fw,
		Introspector: it,
		Rules: []Rule{
			Rule{
				PathMatch: MustCompileMatch("POST:/api/valid/images/<[^/]+>"),
				Scopes:     []string{"ory.valid.images.upload"},
			},
			Rule{
				PathMatch: MustCompileMatch("POST:/api/allowed/<[^/]+>"),
				Scopes:     []string{"ory.allowed.anything"},
				Request: &firewall.TokenAccessRequest{
					Resource: "resource",
					Action: "action",
				},
			},
			Rule{
				PathMatch: MustCompileMatch("POST:/api/dynamic/allowed/<[^/]+>"),
				Scopes:     []string{"ory.allowed.anything"},
				Request: &firewall.TokenAccessRequest{
					Resource: "resource",
					Action: "action",
				},
				RequestFunc: func(r *http.Request) *firewall.TokenAccessRequest {
					return &firewall.TokenAccessRequest{
						Resource: r.Header.Get("resource"),
						Action: "action",
					}
				},
			},
		},
	}

	for k, c := range []struct {
		description string
		r           func() *http.Request
		setup       func()
		isErr       bool
		expect      func(r *http.Request)
	}{
		{
			description: "deny access to a resource if the HTTP method mismatches",
			r: func() *http.Request {
				r, _ := http.NewRequest("GET", "https://foobar/api/valid/images/asdf", nil)
				return r
			},
			isErr: true,
		},
		{
			description: "deny access to a resource if the URL does not have a rule",
			r: func() *http.Request {
				r, _ := http.NewRequest("POST", "https://foobar/api/valid/foobar", nil)
				return r
			},
			isErr: true,
		},
		{
			description: "deny access to a resource if the firewall decides that the token is invalid",
			r: func() *http.Request {
				r, _ := http.NewRequest("POST", "https://foobar/api/valid/images/asdf", nil)
				return r
			},
			setup: func() {
				fw.EXPECT().TokenFromRequest(gomock.Any()).Return("token")
				it.EXPECT().IntrospectToken(gomock.Any(), gomock.Eq("token"), gomock.Eq("ory.valid.images.upload")).Return(nil, errors.New("foo"))
			},
			isErr: true,
		},
		{
			description: "successfully validate a request that does not require a policy check",
			r: func() *http.Request {
				r, _ := http.NewRequest("POST", "https://foobar/api/valid/images/asdf", nil)
				return r
			},
			setup: func() {
				fw.EXPECT().TokenFromRequest(gomock.Any()).Return("token")
				it.EXPECT().IntrospectToken(gomock.Any(), gomock.Eq("token"), gomock.Eq("ory.valid.images.upload")).Return(&oauth2.Introspection{
					Subject: "foobar",
				}, nil)
			},
			expect: func(r *http.Request) {
				assert.Equal(t, r.Header.Get("X-Firewall-Method"), "valid")
				assert.Equal(t, r.Header.Get("X-Firewall-Subject"), "foobar")
			},
		},
		{
			description: "deny access if a request is denied by the policy check",
			r: func() *http.Request {
				r, _ := http.NewRequest("POST", "https://foobar/api/allowed/asdf", nil)
				return r
			},
			setup: func() {
				fw.EXPECT().TokenFromRequest(gomock.Any()).Return("token")
				fw.EXPECT().TokenAllowed(gomock.Any(), gomock.Eq("token"), gomock.Eq(d.Rules[1].Request), gomock.Eq("ory.allowed.anything")).Return(nil, errors.New("foo"))
			},
			isErr: true,
		},
		{
			description: "successfully validate a request and perform a policy check",
			r: func() *http.Request {
				r, _ := http.NewRequest("POST", "https://foobar/api/allowed/asdf", nil)
				return r
			},
			setup: func() {
				fw.EXPECT().TokenFromRequest(gomock.Any()).Return("token")
				fw.EXPECT().TokenAllowed(gomock.Any(), gomock.Eq("token"), gomock.Eq(d.Rules[1].Request), gomock.Eq("ory.allowed.anything")).Return(&firewall.Context{
					Subject: "foobar",
				}, nil)
			},
			expect: func(r *http.Request) {
				assert.Equal(t, r.Header.Get("X-Firewall-Method"), "allowed")
				assert.Equal(t, r.Header.Get("X-Firewall-Subject"), "foobar")
			},
		},
		{
			description: "successfully validate a request and perform a policy check, but with a dynamic rule",
			r: func() *http.Request {
				r, _ := http.NewRequest("POST", "https://foobar/api/dynamic/allowed/asdf", nil)
				r.Header.Add("RequiredResource", "dynamic-resource")
				return r
			},
			setup: func() {
				fw.EXPECT().TokenFromRequest(gomock.Any()).Return("token")
				fw.EXPECT().TokenAllowed(gomock.Any(), gomock.Eq("token"), gomock.Not(d.Rules[2].Request), gomock.Eq("ory.allowed.anything")).Return(&firewall.Context{
					Subject: "foobar",
				}, nil)
			},
			expect: func(r *http.Request) {
				assert.Equal(t, r.Header.Get("X-Firewall-Method"), "allowed")
				assert.Equal(t, r.Header.Get("X-Firewall-Subject"), "foobar")
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d:%s", k, c.description), func(t *testing.T) {
			if c.setup != nil {
				c.setup()
			}

			r := c.r()
			d.Allowed(r)
			err := roundTrip(r)
			if c.isErr {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
			c.expect(r)
		})
	}
}