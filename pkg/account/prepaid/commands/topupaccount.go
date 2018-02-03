package commands

import (
	eh "github.com/looplab/eventhorizon"
	"github.com/richardcase/casecardgo/pkg/account/prepaid"
	"github.com/shopspring/decimal"
)

func init() {
	eh.RegisterCommand(func() eh.Command { return &TopupAccount{} })
}

const TopupAccountCommand = eh.CommandType("prepaid:topup")

var _ = eh.Command(&TopupAccount{})

type TopupAccount struct {
	ID     eh.UUID         `json:"id"`
	Amount decimal.Decimal `json:"amount" bson:"amount"`
}

func (c *TopupAccount) AggregateType() eh.AggregateType { return prepaid.PrePaidAccountAggregateType }
func (c *TopupAccount) AggregateID() eh.UUID            { return c.ID }
func (c *TopupAccount) CommandType() eh.CommandType     { return TopupAccountCommand }
