package director

import (
	"net/http"
	"github.com/ory-am/common/compiler"
	"regexp"
	"github.com/ory-am/hydra/firewall"
	"context"
	"github.com/pkg/errors"
	"fmt"
	"strings"
	"net/url"
	"bytes"
	"io/ioutil"
	"github.com/ory-am/fosite"
	log "github.com/Sirupsen/logrus"
	"github.com/ory-am/hydra/oauth2"
	"github.com/pborman/uuid"
)

var rules = []Rule{
	// developer-ui
	Rule{
		Description: "Reading a subject's images from the vault",
		PathMatch: MustCompileMatch("<GET|POST|OPTIONS>:/developer-ui<.*>"),
		Public: true,
	},
	Rule{
		Description: "Reading a subject's images from the vault",
		PathMatch: MustCompileMatch("<GET|POST|OPTIONS>:/developer-ui-gateway/<graph|auth|auth/callback>"),
		Public: true,
	},

	// plugin-gateway
	Rule{
		Description: "Reading a subject's images from the vault",
		PathMatch: MustCompileMatch("<GET|POST|OPTIONS>:/plugin-gateway/<.*>"),
	},

	// redux-logger
	Rule{
		Description: "Reading a subject's images from the vault",
		PathMatch: MustCompileMatch("<GET|POST|OPTIONS>:/redux-logger/events/<.*>"),
		Public: true,
	},


	// Vault
	Rule{
		Description: "Reading a subject's images from the vault",
		PathMatch: MustCompileMatch("GET:/vault/images/<[^/]+>"),
		Scopes:      []string{"ory.vault.images.read"},
	},
	Rule{
		Description: "Uploading an image to a subject's vault",
		PathMatch: MustCompileMatch("POST:/vault/images/<[^/]+>"),
		Scopes:      []string{"ory.vault.images.write"},
	},

	// Quota
	Rule{
		Description: "Reading the service quota of a subject",
		PathMatch: MustCompileMatch("GET:/quota/quotas/<[^/]+>/<[^/]+>"),
		Scopes:      []string{"ory.quota.quotas.read"},
	},
	Rule{
		Description: "Reading the service quota of a subject",
		PathMatch: MustCompileMatch("POST:/quota/quotas/<[^/]+>/<[^/]+>"),
		Scopes:      []string{"ory.quota.quotas.write"},
	},
	Rule{
		Description: "Reading the service quota of a subject",
		PathMatch: MustCompileMatch("GET:/quota/history/<[^/]+>"),
		Scopes:      []string{"ory.quota.quotas.read"},
	},
	Rule{
		Description: "Reading the service quota of a subject",
		PathMatch: MustCompileMatch("GET:/quota/quotas/<[^/]+>"),
		Scopes:      []string{"ory.quota.quotas.read"},
	},
	Rule{
		Description: "Reading the service quota of a subject",
		PathMatch: MustCompileMatch("POST:/subscriptions/<[^/]+>"),
		Scopes:      []string{"ory.quota.subscriptions.update"},
	},

	// API KEYS
	Rule{
		Description: "Reading a subject's api key",
		PathMatch: MustCompileMatch("GET:/api-keys/keys/<[^/]+>"),
		Scopes:      []string{"ory.api-keys.keys.read"},
	},
	Rule{
		Description: "Reading all api keys of a subject",
		PathMatch: MustCompileMatch("GET:/api-keys/authorize"),
		Public: true,
	},
	Rule{
		Description: "Reading all api keys of a subject",
		PathMatch: MustCompileMatch("GET:/api-keys/authorize/callback"),
		Public: true,
	},
	Rule{
		Description: "Reading all api keys of a subject",
		PathMatch: MustCompileMatch("GET:/api-keys/keys"),
		Scopes:      []string{"ory.api-keys.keys.read"},
	},
	Rule{
		Description: "Updating an existing api key",
		PathMatch: MustCompileMatch("PUT:/api-keys/keys/<[^/]+>"),
		Scopes:      []string{"ory.api-keys.write"},
	},
	Rule{
		Description: "Updating an existing api key",
		PathMatch: MustCompileMatch("POST:/api-keys/keys/<[^/]+>"),
		Scopes:      []string{"ory.api-keys.write"},
	},
	Rule{
		Description: "Delete an existing api key",
		PathMatch: MustCompileMatch("DELETE:/api-keys/keys/<[^/]+>"),
		Scopes:      []string{"ory.api-keys.delete"},
	},
	Rule{
		Description: "Creating a new api key",
		PathMatch: MustCompileMatch("POST:/api-keys/keys"),
		Scopes:      []string{"ory.api-keys.write"},
	},
}

// Rule is a single rule that will get checked on every HTTP request.
type Rule struct {
	// PathMatch is used to match if this rule is responsible for a HTTP request. Rules have the following format:
	//
	// [protocol]:[path]
	//
	// For example:
	//
	// POST:/api/monet/images/<[^/]+>
	//
	// where < and > encapsulate regular expressions. The rule above would therefore match /api/monet/images/foo, /api/monet/images/bar, but not /api/monet/images.
	PathMatch   *regexp.Regexp

	// Scopes: If the path matches, check if the following scopes have been granted.
	Scopes      []string

	// Request can be used to check if the token's subject is allowed to perform an action, using ladon policies.
	Request     *firewall.TokenAccessRequest

	// RequestFunc has the same effect like Request but allows you to orchestrate the request during runtime. This let's you extract additional data from the request, or perform some other magic.
	RequestFunc RequestFunc

	// Public sets if the endpoint is public, thus not needing any authorization at all.
	Public      bool

	Description string
}

func NewDirector(target *url.URL, fw firewall.Firewall, it oauth2.Introspector) *Director {
	return &Director{
		Rules: rules,
		TargetURL: target,
		Firewall: fw,
		Introspector: it,
	}
}

type RequestFunc func(r *http.Request) *firewall.TokenAccessRequest

func MustCompileMatch(path string) *regexp.Regexp {
	reg, err := compiler.CompileRegex(path, '<', '>')
	if err != nil {
		// FIXME panic as long as this is being set up statically #14
		panic(err)
	}
	return reg
}

type Director struct {
	Rules     []Rule
	Firewall  firewall.Firewall
	Introspector oauth2.Introspector
	TargetURL *url.URL
}

type key int

const wasDenied key = 0

func (d *Director) RoundTrip(r *http.Request) (*http.Response, error) {
	if err, ok := r.Context().Value(wasDenied).(error); ok && err != nil {
		he := fosite.ErrorToRFC6749Error(err)
		if he.Name == fosite.UnknownErrorName {
			log.WithError(err).WithField("reason", he.Debug).Print("An unrecognized error occured during access control decision")
		}
		return &http.Response{
			StatusCode: he.StatusCode,
			Body: ioutil.NopCloser(bytes.NewBufferString(he.Description)),
		}, nil
	}

	res, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		log.WithError(err).Print("RoundTrip failed")
		return res, err
	}

	return res, err
}

func (d *Director) Allowed(r *http.Request) {
	targetQuery := d.TargetURL.RawQuery
	r.URL.Scheme = d.TargetURL.Scheme
	r.URL.Host = d.TargetURL.Host
	r.URL.Path = r.URL.Path

	if targetQuery == "" || r.URL.RawQuery == "" {
		r.URL.RawQuery = targetQuery + r.URL.RawQuery
	} else {
		r.URL.RawQuery = targetQuery + "&" + r.URL.RawQuery
	}
	if _, ok := r.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		r.Header.Set("User-Agent", "")
	}
	if _, ok := r.Header["X-Request-Id"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		r.Header.Set("X-Request-Id", uuid.New())
	}

	for _, rule := range d.Rules {
		if !rule.PathMatch.MatchString(fmt.Sprintf("%s:%s", r.Method, r.URL.Path)) {
			continue;
		}

		if rule.Public {
			r.Header.Add("X-Firewall-Method", "anonymous")
			r.Header.Add("X-Firewall-Subject", "anonymous")
			r.Header.Add("X-Firewall-Scopes", strings.Join(rule.Scopes, " "))
			return
		}

		lr := rule.Request
		if rule.RequestFunc != nil {
			lr = rule.RequestFunc(r)
		}

		if lr == nil {
			c, err := d.Introspector.IntrospectToken(context.Background(), d.Firewall.TokenFromRequest(r), rule.Scopes...)
			if err != nil {
				*r = *r.WithContext(context.WithValue(r.Context(), wasDenied, err))
				return
			}

			r.Header.Add("X-Firewall-Method", "valid")
			r.Header.Add("X-Firewall-Subject", c.Subject)
			r.Header.Add("X-Firewall-Scopes", c.Scope)
			return
		}

		c, err := d.Firewall.TokenAllowed(context.Background(), d.Firewall.TokenFromRequest(r), lr, rule.Scopes...)
		if err != nil {
			*r = *r.WithContext(context.WithValue(r.Context(), wasDenied, err))
			return
		}

		r.Header.Add("X-Firewall-Method", "allowed")
		r.Header.Add("X-Firewall-Subject", c.Subject)
		r.Header.Add("X-Firewall-Scopes", strings.Join(c.GrantedScopes, " "))
		return
	}

	*r = *r.WithContext(context.WithValue(r.Context(), wasDenied, errors.Errorf("The endpoint is not protected by the firewall thus access is denied: %s", fmt.Sprintf("%s:%s", r.Method, r.URL.Path))))
	return
}
