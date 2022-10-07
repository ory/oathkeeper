// Copyright © 2022 Ory Corp

package main

// The standard error format
// swagger:model genericError
type genericError struct {
	Code int `json:"code,omitempty"`

	Status string `json:"status,omitempty"`

	Request string `json:"request,omitempty"`

	Reason string `json:"reason,omitempty"`

	Details []map[string]interface{} `json:"details,omitempty"`

	Message string `json:"message"`
}

// An empty response
// swagger:response emptyResponse
type emptyResponse struct{}
