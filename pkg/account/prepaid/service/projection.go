package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/looplab/eventhorizon/eventhandler/projector"

	"github.com/golang/glog"
	eh "github.com/looplab/eventhorizon"
	eventbus "github.com/looplab/eventhorizon/eventbus/local"
	"github.com/looplab/eventhorizon/httputils"
	repo "github.com/looplab/eventhorizon/repo/mongodb"
	evts "github.com/richardcase/casecardgo/pkg/account/prepaid/events"
	"github.com/richardcase/casecardgo/pkg/account/prepaid/projections"
	"github.com/richardcase/casecardgo/pkg/log"
	eventpublisher "github.com/richardcase/casecardgo/pkg/publisher/nats"
)

type ProjectionService struct {
	http.Handler

	CommandHandler eh.CommandHandler
}

func NewProjectionService(
	mongoURL string,
	natsURL string) (*ProjectionService, error) {

	// Create the repo view
	summaryRepo, err := repo.NewRepo(mongoURL, "prepaid", "summary")
	if err != nil {
		return nil, fmt.Errorf("Could not create repo: %s", err)
	}
	summaryRepo.SetEntityFactory(func() eh.Entity { return &projections.PrepaidAccountSummary{} })

	// Create a the bus to distribute events
	eventBus := eventbus.NewEventBus()
	eventPublisher, err := eventpublisher.NewEventPublisher("casecard", natsURL)
	if err != nil {
		return nil, fmt.Errorf("Error creating event publisher: %s", err)
	}
	eventBus.SetPublisher(eventPublisher)

	// Add event logger
	eventPublisher.AddObserver(&log.EventLogger{})

	// Add the projections
	summaryProjector := projector.NewEventHandler(
		projections.NewPrepaidAccountSummaryProjector(),
		summaryRepo)

	summaryProjector.SetEntityFactory(func() eh.Entity { return &projections.PrepaidAccountSummary{} })
	addHandlers(summaryProjector, eventBus)

	// HACK: This is a hack as the projection logic doesn't work
	eventPublisher.AddObserver(&projectorWithObserver{handler: summaryProjector})

	// Setup the HTTP handler
	h := http.NewServeMux()
	h.Handle("/api/prepaid/summary", httputils.QueryHandler(summaryRepo))

	logger := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		glog.Infof("Received HTTP request %s %s", r.Method, r.URL)
		h.ServeHTTP(w, r)
	})

	service := &ProjectionService{
		Handler: logger,
	}

	return service, nil
}

func (s *ProjectionService) Run(listenAddress string) error {

	glog.Info("Started projection service")
	err := http.ListenAndServe(listenAddress, s.Handler)
	if err != nil {
		return fmt.Errorf("Error listening: %s", err.Error())
	}
	glog.Info("Shutting down projection service")

	return nil
}

func addHandlers(projector eh.EventHandler, evtbus eh.EventBus) {
	evtbus.AddHandler(projector, evts.AccountOpened)
	evtbus.AddHandler(projector, evts.AccountToppedUp)
}

// this is a hack as projects don't work
type projectorWithObserver struct {
	handler eh.EventHandler
}

func (p *projectorWithObserver) Notify(ctx context.Context, event eh.Event) {
	err := p.handler.HandleEvent(ctx, event)
	if err != nil {
		glog.Errorf("Error handling event in projection: %s", err.Error())
	}
}
