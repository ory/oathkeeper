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

// Package main ORY Oathkeeper
//
// ORY Oathkeeper is a reverse proxy that checks the HTTP Authorization for validity against a set of rules. This service uses Hydra to validate access tokens and policies.
//
//     Schemes: http, https
//     Host:
//     BasePath: /
//     Version: Latest
//     Contact: ORY <hi@ory.am> https://www.ory.am
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     Extensions:
//     ---
//     x-request-id: string
//     x-forwarded-proto: string
//     ---
//
// swagger:meta
package main
