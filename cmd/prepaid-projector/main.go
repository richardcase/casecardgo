package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/richardcase/casecardgo/pkg/account/prepaid/service"
	"github.com/richardcase/casecardgo/pkg/version"
)

var (
	mongoURL      string
	natsURL       string
	listenAddress string
)

func main() {
	flag.Parse()

	glog.Info("Starting prepaid card projection service.....")
	version.OutputVersion()
	dumpConfig()

	projectionSvc, err := service.NewProjectionService(mongoURL, natsURL)
	if err != nil {
		glog.Fatalf("Error creating projection service: %s", err.Error())
	}

	if err := projectionSvc.Run(listenAddress); err != nil {
		glog.Fatalf("Error running projections service: %s", err.Error())
	}
}

func dumpConfig() {
	glog.Infof("Mongo URL = %s", mongoURL)
	glog.Infof("NATS URL = %s", natsURL)
	glog.Infof("Listen Addr = %s", listenAddress)
}

func init() {
	flag.StringVar(&mongoURL, "mongourl", "localhost:27017", "The mongo URL")
	flag.StringVar(&natsURL, "natsurl", "nats://localhost:4222", "The NATS URL")
	flag.StringVar(&listenAddress, "listen-addr", ":8081", "The address/port to listen on")
}
