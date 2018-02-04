package projections

import (
	"context"
	"fmt"
	"time"

	"github.com/leekchan/accounting"
	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/eventhandler/projector"
	evts "github.com/richardcase/casecardgo/pkg/account/prepaid/events"
)

type PrepaidAccountSummary struct {
	ID      eh.UUID `bson:"_id"`
	Version int

	CardNumber string

	AvailableBalance float64
	BlockedAmount    float64
	TotalLoaded      float64

	Transactions []Transaction

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
		summary.AvailableBalance += data.Amount
		summary.TotalLoaded += data.Amount
		transaction := p.createTransaction(event.Timestamp(), "Topup from "+data.Source, data.Amount)
		summary.Transactions = append(summary.Transactions, transaction)

	case evts.AuthorizationApproved:
		data, ok := event.Data().(*evts.AuthorizationApprovedData)
		if !ok {
			return nil, fmt.Errorf("Projector: invalid event data type: %v", event.Data())
		}
		summary.AvailableBalance -= data.Amount
		summary.BlockedAmount += data.Amount
		transaction := p.createTransaction(event.Timestamp(), "Block from merchant "+data.MerchantId, -data.Amount)
		summary.Transactions = append(summary.Transactions, transaction)

	default:
		return nil, fmt.Errorf("Projector: could not handle event: %s", event.String())
	}

	summary.Version++
	summary.LastActivity = event.Timestamp()

	return summary, nil
}

func (p *PrepaidAccountSummaryProjector) createTransaction(date time.Time, memo string, amount float64) Transaction {
	return Transaction{
		Date:   date,
		Amount: amount,
		Memo:   memo,
	}
}

type Transaction struct {
	Date   time.Time
	Memo   string
	Amount float64
}

func (t Transaction) String() string {
	ac := accounting.Accounting{Symbol: "£", Precision: 2}
	return fmt.Sprintf("%s : %s : %s", t.Date.Local(), t.Memo, ac.FormatMoneyFloat64(t.Amount))
}
