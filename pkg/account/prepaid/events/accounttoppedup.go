package events

import (
	eh "github.com/looplab/eventhorizon"
)

const AccountToppedUp = eh.EventType("prepaid:topped_up")

func init() {
	eh.RegisterEventData(AccountToppedUp, func() eh.EventData {
		return &AccountToppedUpData{}
	})
}

type AccountToppedUpData struct {
	ID     int     `json:"id" bson:"id"`
	Amount float64 `json:"amount" bson:"amount"`
}
