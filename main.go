package main

import (
	"github.com/redhat-nfvpe/service-assurance-poc/amqp"
	"github.com/redhat-nfvpe/service-assurance-poc/cacheutil"
	"github.com/prometheus/client_golang/prometheus"

	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	lastPull = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "serviceassurancecollectd_last_pull_timestamp_seconds",
			Help: "Unix timestamp of the last received collectd metrics pull in seconds.",
		},
	)
)

/*************** HTTP HANDLER***********************/
type cacheHandler struct {
	cache *cacheutil.InputDataV2
}

/*func (h *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
   for hostname,m:= range *h.cache {
    fmt.Fprintln(w, hostname)
		for k :=range m.GetMetrics(hostname){
			fmt.Fprintln(w, k)
		}
	}

}*/
// Describe implements prometheus.Collector.
func (c *cacheHandler) Describe(ch chan<- *prometheus.Desc) {
	ch <- lastPull.Desc()
}

// Collect implements prometheus.Collector.
//need improvement add lock etc etc
func (c *cacheHandler) Collect(ch chan<- prometheus.Metric) {
	log.Println("Collect ")
	ch <- lastPull
	for _, plugin := range c.cache.GetHosts() {
		//fmt.Fprintln(w, hostname)
		plugin.GetNewMetric(ch)
	}

	lastPull.Set(float64(time.Now().UnixNano()) / 1e9)

}

// Usage and command-line flags
func usage() {
	fmt.Fprintln(os.Stderr, `Required commandline argument missing`)
	fmt.Fprintln(os.Stdout, `''`)
	fmt.Fprintln(os.Stdout, `For running with AMQP and Prometheus use following option`)
	fmt.Fprintln(os.Stdout, `'********************* Production *********************'`)
	fmt.Fprintln(os.Stdout, `go run main.go -mhost=localhost -mport=8081 -amqpurl=10.19.110.5:5672/collectd/telemetry `)
	fmt.Fprintln(os.Stdout, `'**************************************************************'`)
	fmt.Fprintln(os.Stdout, `''`)
	fmt.Fprintln(os.Stdout, `''`)
	fmt.Fprintln(os.Stdout, `For running Sample data wihout AMQP use following option\n`)
	fmt.Fprintln(os.Stdout, `'********************* Sample Data *********************'`)
	fmt.Fprintln(os.Stdout, `go run main.go -mhost=localhost -mport=8081 -usesample=true -h=10 -p=100 -t=-1 `)
	fmt.Fprintln(os.Stdout, `'**************************************************************'`)
	flag.PrintDefaults()
}

var fExporterhost = flag.String("mhost", "localhost", "Metrics url for Prometheus to export. ")
var fExporterport = flag.Int("mport", 8081, "Metrics port for Prometheus to export (http://localhost:<port>/metrics) ")
var fAmqpurl = flag.String("amqpurl", "", "AMQP1.0 listener example 127.0.0.1:5672/collectd/telemetry")
var fCount = flag.Int("count", -1, "Stop after receiving this many messages in total(-1 forever) (OPTIONAL)")

var fSampledata = flag.Bool("usesample", false, "Use sample data instead of amqp.This wil not fetch any data from amqp (OPTIONAL)")
var fHosts = flag.Int("h", 1, "No of hosts : Sample hosts required (default 1).")
var fPlugins = flag.Int("p", 100, "No of plugins: Sample plugins per host(default 100).")
var fIterations = flag.Int("t", 1, "No of times to run sample data (default 1) -1 for ever.")

func main() {
	flag.Usage = usage
	flag.Parse()

	if *fSampledata == false && len(*fAmqpurl) == 0 {
		log.Println("AMQP URL is not provided")
		usage()
		os.Exit(1)
	}
	//Cache sever to process and serve the exporter
	var cacheserver = cacheutil.NewCacheServer()

	myHandler := &cacheHandler{cache: cacheserver.GetCache()}

	prometheus.MustRegister(myHandler)
	http.Handle("/metrics", prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Collectd Exporter</title></head>
             <body>
             <h1>Collectd Exporter</h1>
             <p><a href='/metrics'>Metrics</a></p>
             </body>
             </html>`))
	})

	//run exporter fro prometheus to scrape
	go func() {
		var metricsURL = fmt.Sprintf("%s:%d", *fExporterhost, *fExporterport)
		log.Fatal(http.ListenAndServe(metricsURL, nil))
	}()

	if *fSampledata {
		if *fIterations == -1 {
			*fIterations = 9999999
		}
		var hostwaitgroup sync.WaitGroup
		var jsondata = cacheutil.GenerateCollectdJSON("hostname", "pluginname")
		fmt.Printf("Test data  will run for %d times ", *fIterations)
		for times := 1; times <= *fIterations; times++ {
			hostwaitgroup.Add(*fHosts)
			for hosts := 0; hosts < *fHosts; hosts++ {
				go func(host_id int) {
					defer hostwaitgroup.Done()
					var hostname = fmt.Sprintf("%s_%d", "redhat.bosoton.nfv", host_id)
					go cacheutil.GenrateSampleData(hostname, *fPlugins, jsondata, cacheserver)
				}(hosts)

			}
			hostwaitgroup.Wait()
			time.Sleep(time.Second * 1)
		}

	} else {
		//aqp listener if sample is requested then amqp will not be used but random sample data will be used
		notifier := make(chan string) // Channel for messages from goroutines to main()
		var amqpurl = fmt.Sprintf("amqp://%s", *fAmqpurl)
		var amqpServer *amqplistener.AMQPServer
		amqpServer = amqplistener.NewAMQPServer(amqpurl, true, *fCount, notifier)
		for {
			data := <-amqpServer.GetNotifier()
			//fmt.Printf("%v",data)
			c := cacheutil.ParseCollectdJSON(data)
			cacheserver.Put(*c)
		}
	}

}
