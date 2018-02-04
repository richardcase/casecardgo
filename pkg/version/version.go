package version

import "github.com/golang/glog"

// Version is the version number of the initializer
var Version string

// BuildDate is the date the application was built
var BuildDate string

// GitHash is the Git commit hash that was used when building the release
var GitHash string

// OutputVersion writes the version information to the log
func OutputVersion() {
	glog.Infof("Version: %s\n", Version)
	glog.Infof("Buld Date: %s\n", BuildDate)
	glog.Infof("Git Hash: %s\n", GitHash)
}
