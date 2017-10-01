package helper

import "github.com/pkg/errors"

var (
	ErrPublicRule         = errors.New("Rule is public and can not create an access request")
	ErrMissingBearerToken = errors.New("This action requires authorization but no bearer token was given")
)
