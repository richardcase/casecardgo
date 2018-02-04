package commands

import (
	eh "github.com/looplab/eventhorizon"
	"github.com/richardcase/casecardgo/pkg/account/prepaid"
)

func init() {
	eh.RegisterCommand(func() eh.Command { return &AuthorizationRequest{} })
}

const AuthorizationRequestCommand = eh.CommandType("prepaid:authorization_request")

var _ = eh.Command(&AuthorizationRequest{})

type AuthorizationRequest struct {
	ID         eh.UUID `json:"id"`
	Amount     float64 `json:"amount" bson:"amount"`
	MerchantId string  `json:"merchantid" bson:"merchantid"`
}

func (c *AuthorizationRequest) AggregateType() eh.AggregateType {
	return prepaid.PrePaidAccountAggregateType
}
func (c *AuthorizationRequest) AggregateID() eh.UUID        { return c.ID }
func (c *AuthorizationRequest) CommandType() eh.CommandType { return AuthorizationRequestCommand }
