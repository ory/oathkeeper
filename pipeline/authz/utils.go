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

	defer func(originalBody io.Closer) { _ = originalBody.Close() }(r.Body)

	var body bytes.Buffer
	_, err := io.Copy(w, io.TeeReader(r.Body, &body))
	if err != nil {
		return err
	}
	r.Body = io.NopCloser(&body)
	return err
}
