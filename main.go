package main

import (
	"github.com/MakeNowJust/heredoc"

	"flag"
	"fmt"
	"os"
)

/*************** main routine ***********************/
// Usage and command-line flags
func usage() {
	doc := heredoc.Doc(`
******[1 Running Metrics ]**************************************
*For running with config file use                             *
********************* config **********************************
$go run metrics/main.go -config sa.config.json
***************************************************************
* For running with AMQP and Prometheus use following option   *
********************* Command Line ****************************
$ go run metrics/main.go -mhost=localhost -mport=8081 -amqp1MetricURL=10.19.110.5:5672/collectd/telemetry
**************************************************************
* For running Sample data wihout AMQP use following option   *
********************* Sample Data *********************************************************
$ go run metrics/main.go -mhost=localhost -mport=8081 -usesample=true -h=10 -p=100 -t=-1  *
*******************************************************************************************
******[2  Running Events ]******************************
*  For running with config file use                  *
********************* config *************************
$ go run events/main.go -config sa.events.config.json
**********************Command line **********************************
$ go run events/main.go --amqpqEventURL=10.19.110.5:5672/collectd/notify -eshost=http://10.19.110.5:9200
**********************************************************
`)
	fmt.Fprintln(os.Stderr, `Required commandline argument missing`)
	fmt.Fprintln(os.Stdout, doc)
	flag.PrintDefaults()
}

func main() {
	// set flags for parsing options
	flag.Usage = usage
	flag.Parse()
	usage()
	os.Exit(1)

}
