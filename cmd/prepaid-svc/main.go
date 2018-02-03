package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/richardcase/casecardgo/pkg/account/prepaid/service"
	"github.com/richardcase/casecardgo/pkg/signals"
	"github.com/richardcase/casecardgo/pkg/version"
)

var (
	mongoURL      string
	redisServer   string
	listenAddress string
)

func main() {
	flag.Parse()

	glog.Info("Starting prepaid card service.....")
	version.OutputVersion()

	stopCH := signals.SetupSignalHandler()

	prepaidSvc, err := service.NewPrepaidService(mongoURL, redisServer)
	if err != nil {
		glog.Fatalf("Error creating prepaid service: %s", err.Error())
	}

	if err := prepaidSvc.Run(listenAddress, stopCH); err != nil {
		glog.Fatalf("Error running prepaid service: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&mongoURL, "mongourl", "localhost:27017", "The mongo URL")
	flag.StringVar(&redisServer, "redis-server", "localhost:6379", "The redis URL")
	flag.StringVar(&listenAddress, "listen-addr", ":8080", "The address/port to listen on")
}
