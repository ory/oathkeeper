// Copyright © 2022 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package api

// Alive returns an ok status if the instance is ready to handle HTTP requests.
//
// swagger:route GET /health/alive api isInstanceAlive
//
// # Check Alive Status
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
//	Produces:
//	- application/json
//
//	Responses:
//	  200: healthStatus
//	  500: genericError
func swaggerIsInstanceAlive() {}

// Ready returns an ok status if the instance is ready to handle HTTP requests and all ReadyCheckers are ok.
//
// swagger:route GET /health/ready api isInstanceReady
//
// # Check Readiness Status
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
//	Produces:
//	- application/json
//
//	Responses:
//	  200: healthStatus
//	  503: healthNotReadyStatus
func swaggerIsInstanceReady() {}

// Version returns this service's versions.
//
// swagger:route GET /version api getVersion
//
// # Get Service Version
//
// This endpoint returns the service version typically notated using semantic versioning.
//
// If the service supports TLS Edge Termination, this endpoint does not require the
// `X-Forwarded-Proto` header to be set.
//
// Be aware that if you are running multiple nodes of this service, the health status will never
// refer to the cluster state, only to a single instance.
//
//	    Produces:
//	    - application/json
//
//		   Responses:
//				200: version
func swaggerGetVersion() {}
