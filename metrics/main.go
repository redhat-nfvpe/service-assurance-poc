package main

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redhat-nfvpe/service-assurance-poc/amqp"
	"github.com/redhat-nfvpe/service-assurance-poc/cacheutil"
	"github.com/redhat-nfvpe/service-assurance-poc/config"
	"github.com/redhat-nfvpe/service-assurance-poc/incoming"

	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"sync"
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
	$go run metrics/main.go -config sa.metrics.config.json
	**************************************************
	For running with AMQP and Prometheus use following option
	********************* Production *********************
	$go run metrics/main.go -mhost=localhost -mport=8081 -amqp1MetricURL=10.19.110.5:5672/collectd/telemetry
	**************************************************************

	For running Sample data wihout AMQP use following option
	********************* Sample Data *********************
	$go run metrics/main.go -mhost=localhost -mport=8081 -usesample=true -h=10 -p=100 -t=-1
	*************************************************************`)
	fmt.Fprintln(os.Stderr, `Required commandline argument missing`)
	fmt.Fprintln(os.Stdout, doc)
	flag.PrintDefaults()
}

func main() {
	// set flags for parsing options
	flag.Usage = usage
	fConfigLocation := flag.String("config", "", "Path to configuration file(optional).if provided ignores all command line options")
	fIncludeStats := flag.Bool("cpustats", false, "Include cpu usage info in http requests (degrades performance)")
	fExporterhost := flag.String("mhost", "localhost", "Metrics url for Prometheus to export. ")
	fExporterport := flag.Int("mport", 8081, "Metrics port for Prometheus to export (http://localhost:<port>/metrics) ")
	fAMQP1MetricURL := flag.String("amqp1MetricURL", "", "AMQP1.0 metrics listener example 127.0.0.1:5672/collectd/telemetry")
	fCount := flag.Int("count", -1, "Stop after receiving this many messages in total(-1 forever) (OPTIONAL)")

	fSampledata := flag.Bool("usesample", false, "Use sample data instead of amqp.This will not fetch any data from amqp (OPTIONAL)")
	fHosts := flag.Int("h", 1, "No of hosts : Sample hosts required (default 1).")
	fPlugins := flag.Int("p", 100, "No of plugins: Sample plugins per host(default 100).")
	fIterations := flag.Int("t", 1, "No of times to run sample data (default 1) -1 for ever.")

	flag.Parse()
	var serverConfig saconfig.MetricConfiguration
	if len(*fConfigLocation) > 0 { //load configuration
		serverConfig = saconfig.LoadMetricConfig(*fConfigLocation)
	} else {
		serverConfig = saconfig.MetricConfiguration{
			AMQP1MetricURL: *fAMQP1MetricURL,
			CPUStats:       *fIncludeStats,
			Exporterhost:   *fExporterhost,
			Exporterport:   *fExporterport,
			DataCount:      *fCount, //-1 for ever which is default
			UseSample:      *fSampledata,
			Sample: saconfig.SampleDataConfig{
				HostCount:   *fHosts,   //no of host to simulate
				PluginCount: *fPlugins, //No of plugin count per hosts
				DataCount:   *fIterations,
			},
		}

	}

	if serverConfig.UseSample == false && (len(serverConfig.AMQP1MetricURL)==0) {
		log.Println("AMQP1 Metrics URL is required")
		usage()
		os.Exit(1)
	}

	//Cache sever to process and serve the exporter
	cacheServer := cacheutil.NewCacheServer()

	myHandler := &cacheHandler{cache: cacheServer.GetCache()}

	if serverConfig.CPUStats == false {
		// Including these stats kills performance when Prometheus polls with multiple targets
		prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
		prometheus.Unregister(prometheus.NewGoCollector())
	}

	prometheus.MustRegister(myHandler)

	handler := http.NewServeMux()
	handler.Handle("/metrics", prometheus.Handler())
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
                                <head><title>Collectd Exporter</title></head>
                                <body>cacheutil
                                <h1>Collectd Exporter</h1>
                                <p><a href='/metrics'>Metrics</a></p>
                                </body>
                                </html>`))
	})
	// Register pprof handlers
	handler.HandleFunc("/debug/pprof/", pprof.Index)
	handler.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	handler.HandleFunc("/debug/pprof/profile", pprof.Profile)
	handler.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	handler.HandleFunc("/debug/pprof/trace", pprof.Trace)

	//run exporter fro prometheus to scrape
	go func() {
		metricsURL := fmt.Sprintf("%s:%d", serverConfig.Exporterhost, serverConfig.Exporterport)
		log.Fatal(http.ListenAndServe(metricsURL, handler))
	}()
	//if running just samples
	if serverConfig.UseSample {
		if serverConfig.Sample.DataCount == -1 {
			serverConfig.Sample.DataCount = 9999999
		}
		var hostwaitgroup sync.WaitGroup
		fmt.Printf("Test data  will run for %d times ", serverConfig.Sample.DataCount)
		for times := 1; times <= serverConfig.Sample.DataCount; times++ {
			hostwaitgroup.Add(serverConfig.Sample.HostCount)
			for hosts := 0; hosts < serverConfig.Sample.HostCount; hosts++ {
				go func(host_id int) {
					defer hostwaitgroup.Done()
					hostname := fmt.Sprintf("%s_%d", "redhat.boston.nfv", host_id)
					incomingType := incoming.NewInComing(incoming.COLLECTD)
					go cacheServer.GenrateSampleData(hostname, serverConfig.Sample.PluginCount, incomingType)
				}(hosts)

			}
			hostwaitgroup.Wait()
			time.Sleep(time.Second * 1)
		}

	} else {
		//aqp listener if sample is requested then amqp will not be used but random sample data will be used
		metricsNotifier := make(chan string) // Channel for messages from goroutines to main()
		var amqpMetricServer *amqplistener.AMQPServer
		///Metric Listener
		amqpMetricsurl := fmt.Sprintf("amqp://%s", serverConfig.AMQP1MetricURL)
		amqpMetricServer = amqplistener.NewAMQPServer(amqpMetricsurl, true, serverConfig.DataCount, metricsNotifier)

	for {
		 select {
		 case data := <-amqpMetricServer.GetNotifier():
			 //fmt.Printf("%v",data)
			 incomingType := incoming.NewInComing(incoming.COLLECTD)
			 incomingType.ParseInputJSON(data)
			 cacheServer.Put(incomingType)
			 continue
		 default:
			 //no activity
		 }
	 }
 }
	//TO DO: to close cache server on keyboard interrupt

}
