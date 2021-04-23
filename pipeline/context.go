package pipeline

import (
	"net/http"
	"net/url"
)

type Context struct {
}

type ContextRequest struct {
	http.Request
	Header       http.Header  `json:"header"`
	MatchContext MatchContext `json:"match"`
	Method       string
}

type ContextResponse struct {
	URL        string      `json:"url"`
	Header     http.Header `json:"header"`
	Proto      string      `json:"proto"`
	Host       string      `json:"host"`
	RemoteAddr string      `json:"remote_address"`

	RegexpCaptureGroups []string `json:"regexp_capture_groups"`
}

type AuthenticationSession struct {
	Subject      string                 `json:"subject"`
	Extra        map[string]interface{} `json:"extra"`
	Header       http.Header            `json:"header"`
	MatchContext MatchContext           `json:"match_context"`
}

type MatchContext struct {
	RegexpCaptureGroups []string `json:"regexp_capture_groups"`
	URL                 *url.URL `json:"url"`
}
