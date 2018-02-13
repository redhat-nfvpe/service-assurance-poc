package main

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redhat-nfvpe/service-assurance-poc/amqp"
	"github.com/redhat-nfvpe/service-assurance-poc/cacheutil"
	"github.com/redhat-nfvpe/service-assurance-poc/config"

	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	lastPull = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "sa_collectd_last_pull_timestamp_seconds",
			Help: "Unix timestamp of the last received collectd metrics pull in seconds.",
		},
	)
)

/*************** HTTP HANDLER***********************/
type cacheHandler struct {
	cache *cacheutil.IncomingDataCache
}

// Describe implements prometheus.Collector.
func (c *cacheHandler) Describe(ch chan<- *prometheus.Desc) {
	ch <- lastPull.Desc()
}

// Collect implements prometheus.Collector.
//need improvement add lock etc etc
func (c *cacheHandler) Collect(ch chan<- prometheus.Metric) {
	for _, plugin := range c.cache.GetHosts() {
		//fmt.Fprintln(w, hostname)
		plugin.GetNewMetric(ch)
	}

	lastPull.Set(float64(time.Now().UnixNano()) / 1e9)
	ch <- lastPull

	for _, plugin := range c.cache.GetHosts() {
		//fmt.Fprintln(w, hostname)
		plugin.GetNewMetric(ch)
	}
}

/*************** main routine ***********************/
// Usage and command-line flags
func usage() {
	doc := heredoc.Doc(`
  For running with config file use
	********************* config *********************
	$go run events/main.go -config sa.events.config.json
	**************************************************
	For running with AMQP and Prometheus use following option
	********************* Production *********************
	$go run events/main.go -amqp1EventURL=10.19.110.5:5672/collectd/notify
	**************************************************************`)
	fmt.Fprintln(os.Stderr, `Required commandline argument missing`)
	fmt.Fprintln(os.Stdout, doc)
	flag.PrintDefaults()
}

func main() {
	// set flags for parsing options
	flag.Usage = usage
	fConfigLocation := flag.String("config", "", "Path to configuration file(optional).if provided ignores all command line options")

	fAMQP1EventURL := flag.String("amqp1EventURL", "", "AMQP1.0 events listener example 127.0.0.1:5672/collectd/notify")

	flag.Parse()
	var serverConfig saconfig.EventConfiguration
	if len(*fConfigLocation) > 0 { //load configuration
		serverConfig = saconfig.LoadEventConfig(*fConfigLocation)
	} else {
		serverConfig = saconfig.EventConfiguration{
			AMQP1EventURL: *fAMQP1EventURL,
		}

	}

	if len(serverConfig.AMQP1EventURL) == 0 {
		log.Println("AMQP1 Event URL is required")
		usage()
		os.Exit(1)
	}

	eventsNotifier := make(chan string) // Channel for messages from goroutines to main()
	var amqpEventServer *amqplistener.AMQPServer
	///Metric Listener
	amqpEventsurl := fmt.Sprintf("amqp://%s", serverConfig.AMQP1EventURL)
	amqpEventServer = amqplistener.NewAMQPServer(amqpEventsurl, true, -1, eventsNotifier)

	for {
		select {
		case event := <-amqpEventServer.GetNotifier():
			log.Printf("Event occured : %#v\n", event)
			continue
		default:
			//no activity
		}
	}

	//TO DO: to close cache server on keyboard interrupt

}
