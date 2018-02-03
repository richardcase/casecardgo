package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	eh "github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/aggregatestore/events"
	"github.com/looplab/eventhorizon/commandhandler/aggregate"
	"github.com/looplab/eventhorizon/commandhandler/bus"
	eventbus "github.com/looplab/eventhorizon/eventbus/local"
	eventstore "github.com/looplab/eventhorizon/eventstore/mongodb"
	"github.com/looplab/eventhorizon/httputils"
	eventpublisher "github.com/looplab/eventhorizon/publisher/redis"
	"github.com/richardcase/casecardgo/pkg/account/prepaid"
	agg "github.com/richardcase/casecardgo/pkg/account/prepaid/aggregate"
	cmds "github.com/richardcase/casecardgo/pkg/account/prepaid/commands"
	"github.com/richardcase/casecardgo/pkg/log"
)

type PrepaidService struct {
	http.Handler

	CommandHandler eh.CommandHandler
}

func NewPrepaidService(
	mongoURL string,
	redisURL string) (*PrepaidService, error) {

	// Explcitly register aggregate
	eh.RegisterAggregate(func(id eh.UUID) eh.Aggregate {
		return agg.NewPrePaidAccountAggregate(id)
	})

	// Create the event store
	eventStore, err := eventstore.NewEventStore(mongoURL, "casecard")
	if err != nil {
		return nil, fmt.Errorf("Error creating event store: %s", err)
	}

	// Create a the bus to distribute events
	eventBus := eventbus.NewEventBus()
	eventPublisher, err := eventpublisher.NewEventPublisher("casecard", redisURL, "")
	if err != nil {
		return nil, fmt.Errorf("Error creating event publisher: %s", err)
	}
	eventBus.SetPublisher(eventPublisher)

	// Add event logger
	eventPublisher.AddObserver(&log.EventLogger{})

	// Create the command bus.
	commandBus := bus.NewCommandHandler()

	// Create a store for prepaid accounts
	prepaidStore, err := events.NewAggregateStore(eventStore, eventBus)
	if err != nil {
		return nil, fmt.Errorf("Error creating aggregate store: %s", err)
	}

	// Create the command handler and register the commands it handles
	prepaidHandler, err := aggregate.NewCommandHandler(prepaid.PrePaidAccountAggregateType, prepaidStore)
	if err != nil {
		return nil, fmt.Errorf("Error creating command handler: %s", err)
	}

	// Wrap command handler with command logging middleware
	loggingHandler := eh.CommandHandlerFunc(func(ctx context.Context, cmd eh.Command) error {
		glog.V(2).Infof("Command: %#v", cmd)
		return prepaidHandler.HandleCommand(ctx, cmd)
	})

	err = commandBus.SetHandler(loggingHandler, cmds.OpenAccountCommand)
	if err != nil {
		return nil, fmt.Errorf("Unable to set command handler: %s", err)
	}
	err = commandBus.SetHandler(loggingHandler, cmds.TopupAccountCommand)
	if err != nil {
		return nil, fmt.Errorf("Unable to set command handler: %s", err)
	}

	// Setup the HTTP handler
	h := http.NewServeMux()
	h.Handle("/dbg/events/", httputils.EventBusHandler(eventPublisher))
	h.Handle("/api/prepaid/open", httputils.CommandHandler(loggingHandler, cmds.OpenAccountCommand))
	h.Handle("/api/prepaid/topup", httputils.CommandHandler(loggingHandler, cmds.TopupAccountCommand))

	logger := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		glog.Infof("Received HTTP request %s %s", r.Method, r.URL)
		h.ServeHTTP(w, r)
	})

	service := &PrepaidService{
		Handler:        logger,
		CommandHandler: loggingHandler,
	}

	return service, nil
}

func (s *PrepaidService) Run(listenAddress string, stopCh <-chan struct{}) error {

	glog.Info("Started prepaid service")
	err := http.ListenAndServe(listenAddress, s.Handler)
	if err != nil {
		return fmt.Errorf("Error listening: %s", err.Error())
	}
	//<-stopCh
	glog.Info("Shutting down prepaid service")

	return nil
}
