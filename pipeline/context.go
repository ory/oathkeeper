// Copyright 2021 Ory GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
