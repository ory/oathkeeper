// Code generated by go-swagger; DO NOT EDIT.

package api

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewGetRuleParams creates a new GetRuleParams object
// with the default values initialized.
func NewGetRuleParams() *GetRuleParams {
	var ()
	return &GetRuleParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetRuleParamsWithTimeout creates a new GetRuleParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetRuleParamsWithTimeout(timeout time.Duration) *GetRuleParams {
	var ()
	return &GetRuleParams{

		timeout: timeout,
	}
}

// NewGetRuleParamsWithContext creates a new GetRuleParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetRuleParamsWithContext(ctx context.Context) *GetRuleParams {
	var ()
	return &GetRuleParams{

		Context: ctx,
	}
}

// NewGetRuleParamsWithHTTPClient creates a new GetRuleParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetRuleParamsWithHTTPClient(client *http.Client) *GetRuleParams {
	var ()
	return &GetRuleParams{
		HTTPClient: client,
	}
}

/*
GetRuleParams contains all the parameters to send to the API endpoint
for the get rule operation typically these are written to a http.Request
*/
type GetRuleParams struct {

	/*ID*/
	ID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get rule params
func (o *GetRuleParams) WithTimeout(timeout time.Duration) *GetRuleParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get rule params
func (o *GetRuleParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get rule params
func (o *GetRuleParams) WithContext(ctx context.Context) *GetRuleParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get rule params
func (o *GetRuleParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get rule params
func (o *GetRuleParams) WithHTTPClient(client *http.Client) *GetRuleParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get rule params
func (o *GetRuleParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithID adds the id to the get rule params
func (o *GetRuleParams) WithID(id string) *GetRuleParams {
	o.SetID(id)
	return o
}

// SetID adds the id to the get rule params
func (o *GetRuleParams) SetID(id string) {
	o.ID = id
}

// WriteToRequest writes these params to a swagger request
func (o *GetRuleParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param id
	if err := r.SetPathParam("id", o.ID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
