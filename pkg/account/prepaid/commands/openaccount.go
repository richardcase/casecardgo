package commands

import (
	eh "github.com/looplab/eventhorizon"
	"github.com/richardcase/casecardgo/pkg/account/prepaid"
)

func init() {
	eh.RegisterCommand(func() eh.Command { return &OpenAccount{} })
}

const OpenAccountCommand = eh.CommandType("prepaid:open")

var _ = eh.Command(&OpenAccount{})

type OpenAccount struct {
	ID            eh.UUID `json:"id"`
	AccountHolder string  `json:"accountholder" bson:"accountholder"`
	Address       string  `json:"address" bson:"address"`
}

func (c *OpenAccount) AggregateType() eh.AggregateType { return prepaid.PrePaidAccountAggregateType }
func (c *OpenAccount) AggregateID() eh.UUID            { return c.ID }
func (c *OpenAccount) CommandType() eh.CommandType     { return OpenAccountCommand }
