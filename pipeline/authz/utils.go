// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package authz

import (
	"bytes"
	"io"
	"net/http"
)

func pipeRequestBody(r *http.Request, w io.Writer) error {
	if r.Body == nil {
		return nil
	}

	var body bytes.Buffer
	defer r.Body.Close() //nolint:errcheck
	_, err := io.Copy(w, io.TeeReader(r.Body, &body))
	if err != nil {
		return err
	}
	r.Body = io.NopCloser(&body)
	return err
}
