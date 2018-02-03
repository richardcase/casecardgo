package aggregate

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dmgk/faker"
	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/aggregatestore/events"

	"github.com/richardcase/casecardgo/pkg/account/prepaid"
	cmds "github.com/richardcase/casecardgo/pkg/account/prepaid/commands"
	evts "github.com/richardcase/casecardgo/pkg/account/prepaid/events"
)

var ErrUnknownCommand = errors.New("Unknown command")

/*func init() {
	eh.RegisterAggregate(func(id eh.UUID) eh.Aggregate {
		return NewPrePaidAccountAggregate(id)
	})
}*/

type PrePaidAccountAggregate struct {
	*events.AggregateBase

	accoundHolder string
	cardNumber    string
	openedOn      time.Time

	available float64
	blocked   float64

	//TODO: finish this add blocks
}

var _ = eh.Aggregate(&PrePaidAccountAggregate{})

func NewPrePaidAccountAggregate(id eh.UUID) *PrePaidAccountAggregate {
	return &PrePaidAccountAggregate{
		AggregateBase: events.NewAggregateBase(prepaid.PrePaidAccountAggregateType, id),
	}
}

func (a *PrePaidAccountAggregate) HandleCommand(ctx context.Context, cmd eh.Command) error {
	switch cmd := cmd.(type) {
	case *cmds.OpenAccount:
		return a.handleAccountOpen(cmd)
	case *cmds.TopupAccount:
		return a.handleTopup(cmd)
	}
	return ErrUnknownCommand
}

func (a *PrePaidAccountAggregate) ApplyEvent(ctc context.Context, event eh.Event) error {
	switch event.EventType() {
	case evts.AccountOpened:
		if data, ok := event.Data().(*evts.AccountOpenedData); ok {
			a.accoundHolder = data.AccountHolder
			a.cardNumber = data.CardNumber
			a.openedOn = data.OpenedOn
		} else {
			return fmt.Errorf("invalid event data type: %s", event.Data())
		}
	case evts.AccountToppedUp:
		if data, ok := event.Data().(*evts.AccountToppedUpData); ok {
			a.available += data.Amount
		} else {
			return fmt.Errorf("invalid event data type: %s", event.Data())
		}
	}
	return nil
}

func (a *PrePaidAccountAggregate) handleAccountOpen(cmd *cmds.OpenAccount) error {
	if cmd.AccountHolder == "" {
		return fmt.Errorf("AccountHolder must be specified")
	}
	a.StoreEvent(evts.AccountOpened,
		&evts.AccountOpenedData{
			AccountHolder: cmd.AccountHolder,
			CardNumber:    faker.Business().CreditCardNumber(),
			OpenedOn:      time.Now(),
		}, time.Now(),
	)

	return nil
}

func (a *PrePaidAccountAggregate) handleTopup(cmd *cmds.TopupAccount) error {
	if cmd.Amount <= 0.0 {
		return fmt.Errorf("Topup must be for a positive amount")
	}

	a.StoreEvent(evts.AccountToppedUp,
		&evts.AccountToppedUpData{
			Amount: cmd.Amount,
		}, time.Now(),
	)

	return nil
}
