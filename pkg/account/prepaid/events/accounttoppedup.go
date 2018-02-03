package events

import (
	eh "github.com/looplab/eventhorizon"
	"github.com/shopspring/decimal"
)

const AccountToppedUp = eh.EventType("prepaid:topped_up")

func init() {
	eh.RegisterEventData(AccountToppedUp, func() eh.EventData {
		return &AccountToppedUpData{}
	})
}

type AccountToppedUpData struct {
	ID     int             `json:"id" bson:"id"`
	Amount decimal.Decimal `json:"amount" bson:"amount"`
}
