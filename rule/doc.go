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

// Package rule implements management capabilities for rules
//
// A rule is used to decide what to do with requests that are hitting the ORY Oathkeeper proxy server. A rule must
// define the HTTP methods and the URL under which it will apply. A URL may not have more than one rule. If a URL
// has no rule applied, the proxy server will return a 404 not found error.
//
// ORY Oathkeeper stores as many rules as required and iterates through them on every request. Rules are essential
// to the way ORY Oathkeeper works.
package rule
