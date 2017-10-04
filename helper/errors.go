package helper

import "github.com/pkg/errors"

var (
	ErrMissingBearerToken     = errors.New("This action requires authorization but no bearer token was given")
	ErrForbidden              = errors.New("Access credentials are not sufficient to access this resource")
	ErrMatchesMoreThanOneRule = errors.New("Expected exactly one rule but found multiple rules")
	ErrMatchesNoRule          = errors.New("Requested url does not match any rules")
	ErrResourceNotFound       = errors.New("The requested resource could not be found")
)
