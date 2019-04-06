// Code generated by go-swagger; DO NOT EDIT.

package rule

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/swag"

	strfmt "github.com/go-openapi/strfmt"
)

// NewListRulesParams creates a new ListRulesParams object
// with the default values initialized.
func NewListRulesParams() *ListRulesParams {
	var ()
	return &ListRulesParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewListRulesParamsWithTimeout creates a new ListRulesParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewListRulesParamsWithTimeout(timeout time.Duration) *ListRulesParams {
	var ()
	return &ListRulesParams{

		timeout: timeout,
	}
}

// NewListRulesParamsWithContext creates a new ListRulesParams object
// with the default values initialized, and the ability to set a context for a request
func NewListRulesParamsWithContext(ctx context.Context) *ListRulesParams {
	var ()
	return &ListRulesParams{

		Context: ctx,
	}
}

// NewListRulesParamsWithHTTPClient creates a new ListRulesParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewListRulesParamsWithHTTPClient(client *http.Client) *ListRulesParams {
	var ()
	return &ListRulesParams{
		HTTPClient: client,
	}
}

/*ListRulesParams contains all the parameters to send to the API endpoint
for the list rules operation typically these are written to a http.Request
*/
type ListRulesParams struct {

	/*Limit
	  The maximum amount of rules returned.

	*/
	Limit *int64
	/*Offset
	  The offset from where to start looking.

	*/
	Offset *int64

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the list rules params
func (o *ListRulesParams) WithTimeout(timeout time.Duration) *ListRulesParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list rules params
func (o *ListRulesParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list rules params
func (o *ListRulesParams) WithContext(ctx context.Context) *ListRulesParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list rules params
func (o *ListRulesParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the list rules params
func (o *ListRulesParams) WithHTTPClient(client *http.Client) *ListRulesParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the list rules params
func (o *ListRulesParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithLimit adds the limit to the list rules params
func (o *ListRulesParams) WithLimit(limit *int64) *ListRulesParams {
	o.SetLimit(limit)
	return o
}

// SetLimit adds the limit to the list rules params
func (o *ListRulesParams) SetLimit(limit *int64) {
	o.Limit = limit
}

// WithOffset adds the offset to the list rules params
func (o *ListRulesParams) WithOffset(offset *int64) *ListRulesParams {
	o.SetOffset(offset)
	return o
}

// SetOffset adds the offset to the list rules params
func (o *ListRulesParams) SetOffset(offset *int64) {
	o.Offset = offset
}

// WriteToRequest writes these params to a swagger request
func (o *ListRulesParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Limit != nil {

		// query param limit
		var qrLimit int64
		if o.Limit != nil {
			qrLimit = *o.Limit
		}
		qLimit := swag.FormatInt64(qrLimit)
		if qLimit != "" {
			if err := r.SetQueryParam("limit", qLimit); err != nil {
				return err
			}
		}

	}

	if o.Offset != nil {

		// query param offset
		var qrOffset int64
		if o.Offset != nil {
			qrOffset = *o.Offset
		}
		qOffset := swag.FormatInt64(qrOffset)
		if qOffset != "" {
			if err := r.SetQueryParam("offset", qOffset); err != nil {
				return err
			}
		}

	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
