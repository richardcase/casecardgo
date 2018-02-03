package log

import (
	"context"

	"github.com/golang/glog"
	eh "github.com/looplab/eventhorizon"
)

type EventLogger struct{}

func (l *EventLogger) Notify(ctx context.Context, event eh.Event) {
	glog.V(2).Infof("Event: %v", event)
}
