package middleware

import (
	"github.com/ory/oathkeeper/driver"
)

// WithRegistry is a test option that writes the internal registry used in the
// middleware to the given pointer. The pointer must be allocaded by the caller
// beforehand (e.g., `regPtr := new(driver.Registry)`).
func WithRegistry(r *driver.Registry) Option {
	return func(o *options) { o.registryAddr = r }
}
