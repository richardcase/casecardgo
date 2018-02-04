package commands

import (
	eh "github.com/looplab/eventhorizon"
	"github.com/richardcase/casecardgo/pkg/account/prepaid"
)

func init() {
	eh.RegisterCommand(func() eh.Command { return &TopupAccount{} })
}

const TopupAccountCommand = eh.CommandType("prepaid:topup")

var _ = eh.Command(&TopupAccount{})

type TopupAccount struct {
	ID     eh.UUID `json:"id"`
	Amount float64 `json:"amount" bson:"amount"`
	Source string  `json:"source" bson:"source"`
}

func (c *TopupAccount) AggregateType() eh.AggregateType { return prepaid.PrePaidAccountAggregateType }
func (c *TopupAccount) AggregateID() eh.UUID            { return c.ID }
func (c *TopupAccount) CommandType() eh.CommandType     { return TopupAccountCommand }
