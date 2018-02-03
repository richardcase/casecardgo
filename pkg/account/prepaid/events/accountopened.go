package events

import (
	"time"

	eh "github.com/looplab/eventhorizon"
)

const AccountOpened = eh.EventType("prepaid:account_opened")

func init() {
	eh.RegisterEventData(AccountOpened, func() eh.EventData {
		return &AccountOpenedData{}
	})
}

type AccountOpenedData struct {
	ID            int       `json:"id" bson:"id"`
	AccountHolder string    `json:"accountholder" bson:"accountholder"`
	CardNumber    string    `json:"cardnumber" bson:"cardnumber"`
	OpenedOn      time.Time `json:"openedon" bson:"openedon"`
}
