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

package api

// Alive returns an ok status if the instance is ready to handle HTTP requests.
//
// swagger:route GET /health/alive api isInstanceAlive
//
// Check alive status
//
// This endpoint returns a 200 status code when the HTTP server is up running.
// This status does currently not include checks whether the database connection is working.
//
// If the service supports TLS Edge Termination, this endpoint does not require the
// `X-Forwarded-Proto` header to be set.
//
// Be aware that if you are running multiple nodes of this service, the health status will never
// refer to the cluster state, only to a single instance.
//
//     Produces:
//     - application/json
//
//     Responses:
//       200: healthStatus
//       500: genericError
func swaggerIsInstanceAlive() {}

// Ready returns an ok status if the instance is ready to handle HTTP requests and all ReadyCheckers are ok.
//
// swagger:route GET /health/ready api isInstanceReady
//
// Check readiness status
//
// This endpoint returns a 200 status code when the HTTP server is up running and the environment dependencies (e.g.
// the database) are responsive as well.
//
// If the service supports TLS Edge Termination, this endpoint does not require the
// `X-Forwarded-Proto` header to be set.
//
// Be aware that if you are running multiple nodes of this service, the health status will never
// refer to the cluster state, only to a single instance.
//
//     Produces:
//     - application/json
//
//     Responses:
//       200: healthStatus
//       503: healthNotReadyStatus
func swaggerIsInstanceReady() {}

// Version returns this service's versions.
//
// swagger:route GET /version api getVersion
//
// Get service version
//
// This endpoint returns the service version typically notated using semantic versioning.
//
// If the service supports TLS Edge Termination, this endpoint does not require the
// `X-Forwarded-Proto` header to be set.
//
// Be aware that if you are running multiple nodes of this service, the health status will never
// refer to the cluster state, only to a single instance.
//
//     Produces:
//     - application/json
//
//	   Responses:
// 			200: version
func swaggerGetVersion() {}
