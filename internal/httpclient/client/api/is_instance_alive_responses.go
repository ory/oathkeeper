// Code generated by go-swagger; DO NOT EDIT.

package api

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/ory/oathkeeper/internal/httpclient/models"
)

// IsInstanceAliveReader is a Reader for the IsInstanceAlive structure.
type IsInstanceAliveReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *IsInstanceAliveReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewIsInstanceAliveOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 500:
		result := NewIsInstanceAliveInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewIsInstanceAliveOK creates a IsInstanceAliveOK with default headers values
func NewIsInstanceAliveOK() *IsInstanceAliveOK {
	return &IsInstanceAliveOK{}
}

/*
IsInstanceAliveOK handles this case with default header values.

healthStatus
*/
type IsInstanceAliveOK struct {
	Payload *models.HealthStatus
}

func (o *IsInstanceAliveOK) Error() string {
	return fmt.Sprintf("[GET /health/alive][%d] isInstanceAliveOK  %+v", 200, o.Payload)
}

func (o *IsInstanceAliveOK) GetPayload() *models.HealthStatus {
	return o.Payload
}

func (o *IsInstanceAliveOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.HealthStatus)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewIsInstanceAliveInternalServerError creates a IsInstanceAliveInternalServerError with default headers values
func NewIsInstanceAliveInternalServerError() *IsInstanceAliveInternalServerError {
	return &IsInstanceAliveInternalServerError{}
}

/*
IsInstanceAliveInternalServerError handles this case with default header values.

genericError
*/
type IsInstanceAliveInternalServerError struct {
	Payload *models.GenericError
}

func (o *IsInstanceAliveInternalServerError) Error() string {
	return fmt.Sprintf("[GET /health/alive][%d] isInstanceAliveInternalServerError  %+v", 500, o.Payload)
}

func (o *IsInstanceAliveInternalServerError) GetPayload() *models.GenericError {
	return o.Payload
}

func (o *IsInstanceAliveInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GenericError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
