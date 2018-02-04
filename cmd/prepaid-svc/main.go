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
	natsURL       string
	listenAddress string
)

func main() {
	flag.Parse()

	glog.Info("Starting prepaid card service.....")
	version.OutputVersion()
	dumpConfig()

	stopCH := signals.SetupSignalHandler()

	prepaidSvc, err := service.NewPrepaidService(mongoURL, natsURL)
	if err != nil {
		glog.Fatalf("Error creating prepaid service: %s", err.Error())
	}

	if err := prepaidSvc.Run(listenAddress, stopCH); err != nil {
		glog.Fatalf("Error running prepaid service: %s", err.Error())
	}
}

func dumpConfig() {
	glog.V(2).Infof("Mongo URL = %s", mongoURL)
	glog.V(2).Infof("NATS URL = %s", natsURL)
	glog.V(2).Infof("Listen Addr = %s", listenAddress)
}

func init() {
	flag.StringVar(&mongoURL, "mongourl", "localhost:27017", "The mongo URL")
	flag.StringVar(&natsURL, "natsurl", "nats://localhost:4222", "The NATS URL")
	flag.StringVar(&listenAddress, "listen-addr", ":8080", "The address/port to listen on")
}
