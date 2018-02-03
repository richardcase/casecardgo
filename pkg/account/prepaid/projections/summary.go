package projections

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/eventhandler/projector"
	evts "github.com/richardcase/casecardgo/pkg/account/prepaid/events"
)

type PrepaidAccountSummary struct {
	ID      eh.UUID `bson:"_id"`
	Version int

	CardNumber string

	AvailableBalance decimal.Decimal
	BlockedAmount    decimal.Decimal
	TotalLoaded      decimal.Decimal

	LastActivity time.Time
}

var _ = eh.Entity(&PrepaidAccountSummary{})
var _ = eh.Versionable(&PrepaidAccountSummary{})

func (p *PrepaidAccountSummary) EntityID() eh.UUID {
	return p.ID
}

func (p *PrepaidAccountSummary) AggregateVersion() int {
	return p.Version
}

type PrepaidAccountSummaryProjector struct{}

func NewPrepaidAccountSummaryProjector() *PrepaidAccountSummaryProjector {
	return &PrepaidAccountSummaryProjector{}
}

func (p *PrepaidAccountSummaryProjector) ProjectorType() projector.Type {
	return projector.Type("PrepaidAccountSummaryProjector")
}

func (p *PrepaidAccountSummaryProjector) Project(ctx context.Context, event eh.Event, entity eh.Entity) (eh.Entity, error) {
	summary, ok := entity.(*PrepaidAccountSummary)
	if !ok {
		return nil, fmt.Errorf("Unable to convert model to PrepaidAccountSummary")
	}

	// Apply events
	switch event.EventType() {
	case evts.AccountOpened:
		data, ok := event.Data().(*evts.AccountOpenedData)
		if !ok {
			return nil, fmt.Errorf("Projector: invalid event data type: %v", event.Data())
		}
		summary.ID = event.AggregateID()
		summary.CardNumber = data.CardNumber

	case evts.AccountToppedUp:
		data, ok := event.Data().(*evts.AccountToppedUpData)
		if !ok {
			return nil, fmt.Errorf("Projector: invalid event data type: %v", event.Data())
		}
		summary.AvailableBalance.Add(data.Amount)
		summary.TotalLoaded.Add(data.Amount)

	default:
		return nil, fmt.Errorf("Projector: could not handle event: %s", event.String())
	}

	summary.Version++
	summary.LastActivity = event.Timestamp()

	return summary, nil
}
