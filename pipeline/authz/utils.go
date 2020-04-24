package authz

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

func pipeRequestBody(r *http.Request, w io.Writer) error {
	if r.Body == nil {
		return nil
	}

	var body bytes.Buffer
	_, err := io.Copy(w, io.TeeReader(r.Body, &body))
	r.Body = ioutil.NopCloser(&body)
	return err
}
