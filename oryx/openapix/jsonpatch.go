// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package openapix

// A JSONPatchDocument request
//
// swagger:model jsonPatchDocument
type JSONPatchDocument []JSONPatch

// A JSONPatch document as defined by RFC 6902
//
// swagger:model jsonPatch
type JSONPatch struct {
	// The operation to be performed. One of "add", "remove", or "replace".
	//
	// required: true
	// example: replace
	Op string `json:"op"`

	// The path to the target path. Uses JSON pointer notation.
	//
	// Learn more [about JSON Pointers](https://datatracker.ietf.org/doc/html/rfc6901#section-5).
	//
	// required: true
	// example: /name
	Path string `json:"path"`

	// The value to be used within the operations.
	//
	// Learn more [about JSON Pointers](https://datatracker.ietf.org/doc/html/rfc6901#section-5).
	//
	// example: foobar
	Value interface{} `json:"value"`

	// The source path for operations that require it. Uses JSON pointer notation.
	// Not used by the currently supported operations ("add", "remove", "replace").
	//
	// Learn more [about JSON Pointers](https://datatracker.ietf.org/doc/html/rfc6901#section-5).
	//
	// example: /name
	From string `json:"from"`
}
