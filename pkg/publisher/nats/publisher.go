package nats

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang/glog"
	nats "github.com/nats-io/go-nats"
	"gopkg.in/mgo.v2/bson"

	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/publisher/local"
)

var ErrCouldNotMarshalEvent = errors.New("could not marshal event")
var ErrCouldNotUnmarshalEvent = errors.New("could not unmarshal event")

type EventPublisher struct {
	*local.EventPublisher

	connection   *nats.Conn
	subscription *nats.Subscription
	subject      string
	ready        chan bool
	exit         chan bool
	errCh        chan Error
}

func NewEventPublisher(appID string, servers string) (*EventPublisher, error) {
	nc, err := nats.Connect(servers)
	if err != nil {
		return nil, fmt.Errorf("Enable to connect to NATS: %s", err.Error())
	}

	// Create the subject
	subject := appID + "_events"

	pub := &EventPublisher{
		EventPublisher: local.NewEventPublisher(),
		connection:     nc,
		subject:        subject,
		ready:          make(chan bool, 1),
		exit:           make(chan bool),
		errCh:          make(chan Error, 20),
	}

	go pub.recv()

	return pub, nil

}

func (p *EventPublisher) PublishEvent(ctx context.Context, event eh.Event) error {

	natsEvent := natsEvent{
		AggregateID:   event.AggregateID(),
		AggregateType: event.AggregateType(),
		EventType:     event.EventType(),
		Version:       event.Version(),
		Timestamp:     event.Timestamp(),
		Context:       eh.MarshalContext(ctx),
	}

	if event.Data() != nil {
		rawData, err := bson.Marshal(event.Data())
		if err != nil {
			return ErrCouldNotMarshalEvent
		}
		natsEvent.RawData = bson.Raw{Kind: 3, Data: rawData}
	}

	var data []byte
	var err error
	if data, err = bson.Marshal(natsEvent); err != nil {
		return ErrCouldNotMarshalEvent
	}

	err = p.connection.Publish(p.subject, data)
	if err != nil {
		return fmt.Errorf("Error publishing event: %s", err.Error())
	}

	return nil
}

func (p *EventPublisher) Close() error {
	select {
	case p.exit <- true:
	default:
		glog.Info("NATS Publisher: already closed")
	}
	<-p.exit

	p.connection.Close()

	return nil
}

func (p *EventPublisher) recv() {
	select {
	case p.ready <- true:
	default:
	}

	sub, err := p.connection.Subscribe(p.subject, func(msg *nats.Msg) {
		if err := p.handleMessage(msg); err != nil {
			glog.Errorf("NATS Publisher: Error publishing: %s", err.Error())
		}
	})
	if err != nil {
		glog.Errorf("NATS Publisher: Error creating subscription: %s", err.Error())
		return
	}

	glog.Info("NATS Publisher: Starting receiving")
	go func() {
		<-p.exit
		err := sub.Unsubscribe()
		if err != nil {
			glog.Errorf("NATS Publisher: Error unsubscribing: %s", err.Error())
		}
		close(p.exit)
	}()
}

func (p *EventPublisher) handleMessage(msg *nats.Msg) error {
	data := bson.Raw{
		Kind: 3,
		Data: msg.Data,
	}
	var natsevent natsEvent
	if err := data.Unmarshal(&natsevent); err != nil {
		return ErrCouldNotUnmarshalEvent
	}

	if data, err := eh.CreateEventData(natsevent.EventType); err == nil {
		if err := natsevent.RawData.Unmarshal(data); err != nil {
			return ErrCouldNotUnmarshalEvent
		}

		natsevent.data = data
		natsevent.RawData = bson.Raw{}
	}

	event := event{natsEvent: natsevent}
	ctx := eh.UnmarshalContext(natsevent.Context)

	return p.EventPublisher.PublishEvent(ctx, event)
}

type natsEvent struct {
	EventType     eh.EventType           `bson:"event_type"`
	RawData       bson.Raw               `bson:"data,omitempty"`
	data          eh.EventData           `bson:"-"`
	Timestamp     time.Time              `bson:"timestamp"`
	AggregateType eh.AggregateType       `bson:"aggregate_type"`
	AggregateID   eh.UUID                `bson:"_id"`
	Version       int                    `bson:"version"`
	Context       map[string]interface{} `bson:"context"`
}

type event struct {
	natsEvent
}

func (e event) EventType() eh.EventType {
	return e.natsEvent.EventType
}

func (e event) Data() eh.EventData {
	return e.natsEvent.data
}

func (e event) Timestamp() time.Time {
	return e.natsEvent.Timestamp
}

func (e event) AggregateType() eh.AggregateType {
	return e.natsEvent.AggregateType
}

func (e event) AggregateID() eh.UUID {
	return e.natsEvent.AggregateID
}

func (e event) Version() int {
	return e.natsEvent.Version
}

func (e event) String() string {
	return fmt.Sprintf("%s@%d", e.natsEvent.EventType, e.natsEvent.Version)
}

type Error struct {
	Err   error
	Ctx   context.Context
	Event eh.Event
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Event.String(), e.Err.Error())
}
