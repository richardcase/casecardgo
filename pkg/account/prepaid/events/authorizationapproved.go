package events

import (
	eh "github.com/looplab/eventhorizon"
)

const AuthorizationApproved = eh.EventType("prepaid:authorization_approved")

func init() {
	eh.RegisterEventData(AuthorizationApproved, func() eh.EventData {
		return &AuthorizationApprovedData{}
	})
}

type AuthorizationApprovedData struct {
	Amount          float64 `json:"amount" bson:"amount"`
	MerchantId      string  `json:"merchantid" bson:"merchantid"`
	AuthorizationId eh.UUID `json:"authorizationid"`
}
