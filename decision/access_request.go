package decision

import "github.com/ory/hydra/sdk/go/hydra/swagger"

type AccessRequest struct {
	swagger.WardenTokenAccessRequest
	Public bool
}
