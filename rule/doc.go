// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

// Package rule implements management capabilities for rules
//
// A rule is used to decide what to do with requests that are hitting the ORY Oathkeeper proxy server. A rule must
// define the HTTP methods and the URL under which it will apply. A URL may not have more than one rule. If a URL
// has no rule applied, the proxy server will return a 404 not found error.
//
// ORY Oathkeeper stores as many rules as required and iterates through them on every request. Rules are essential
// to the way ORY Oathkeeper works.
package rule
