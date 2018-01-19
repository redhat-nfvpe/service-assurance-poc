package main

import (
	"github.com/aneeshkp/service-assurance-goclient/amqp"
	"github.com/aneeshkp/service-assurance-goclient/cacheutil"
	"github.com/prometheus/client_golang/prometheus"

	"net/http"
	"sync"
	"time"
	"fmt"
	"os"
	"flag"
	"log"
)

var (
	lastPull = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "service_assurance_collectd_last_pull_timestamp_seconds",
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
	lastPull.Set(float64(time.Now().UnixNano()) / 1e9)
	ch <- lastPull

	for _, plugin := range c.cache.GetHosts() {
		//fmt.Fprintln(w, hostname)
		plugin.GetNewMetric(ch)
	}
}

// Usage and command-line flags
func usage() {
	fmt.Fprintln(os.Stderr, `Required commandline argument missing`)
	fmt.Fprintln(os.Stdout,`''`)
	fmt.Fprintln(os.Stdout, `For running with AMQP and Prometheus use following option`)
	fmt.Fprintln(os.Stdout,`'********************* Production *********************'`)
	fmt.Fprintln(os.Stdout, `go run main.go -mhost=localhost -mport=8081 -amqpurl=10.19.110.5:5672/collectd/telemetry `)
	fmt.Fprintln(os.Stdout,`'**************************************************************'`)
  fmt.Fprintln(os.Stdout,`''`)
  fmt.Fprintln(os.Stdout,`''`)
	fmt.Fprintln(os.Stdout, `For running Sample data wihout AMQP use following option\n`)
	fmt.Fprintln(os.Stdout,`'********************* Sample Data *********************'`)
	fmt.Fprintln(os.Stdout, `go run main.go -mhost=localhost -mport=8081 -usesample=true -h=10 -p=100 -t=-1 `)
	fmt.Fprintln(os.Stdout,`'**************************************************************'`)
	flag.PrintDefaults()
}


var f_exporterhost = flag.String("mhost", "localhost", "Metrics url for Prometheus to export. ")
var f_exporterport = flag.Int("mport", 8081, "Metrics port for Prometheus to export (http://localhost:<port>/metrics) ")
var f_amqpurl = flag.String("amqpurl", "", "AMQP1.0 listener example 127.0.0.1:5672/collectd/telemetry")
var f_count = flag.Int("count", -1, "Stop after receiving this many messages in total(-1 forever) (OPTIONAL)")

var f_sampledata = flag.Bool("usesample", false, "Use sample data instead of amqp.This wil not fetch any data from amqp (OPTIONAL)")
var f_hosts = flag.Int("h", 1, "No of hosts : Sample hosts required (deafult 1).")
var f_plugins = flag.Int("p", 100, "No of plugins: Sample plugins per host(default 100).")
var f_iterations = flag.Int("t", 1, "No of times to run sample data (default 1) -1 for ever.")


func main() {
	flag.Usage = usage
  flag.Parse()

	if *f_sampledata==false && len(*f_amqpurl) == 0 {
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
		var metricsURL=fmt.Sprintf("%s:%d",*f_exporterhost,*f_exporterport)
		log.Fatal(http.ListenAndServe(metricsURL, nil))
	}()

  if *f_sampledata {
		  if *f_iterations==-1{
				*f_iterations=9999999
			}
			var hostwaitgroup sync.WaitGroup
			var jsondata = cacheutil.GenerateCollectdJson("hostname", "pluginname")
			fmt.Printf("Test data  will run for %d times ",*f_iterations)
			for times := 1; times <= *f_iterations; times++ {
				hostwaitgroup.Add(*f_hosts)
				for hosts := 0; hosts < *f_hosts; hosts++ {
					go func(host_id int) {
						defer hostwaitgroup.Done()
						var hostname = fmt.Sprintf("%s_%d", "redhat.bosoton.nfv", host_id)
						go cacheutil.GenrateSampleData(hostname, *f_plugins, jsondata, cacheserver)
					}(hosts)

				}
				hostwaitgroup.Wait()
				time.Sleep(time.Second * 1)
			}

	}else{
		//aqp listener if sample is requested then amqp will not be used but random sample data will be used
		notifier := make(chan string) // Channel for messages from goroutines to main()
		var amqpurl = fmt.Sprintf("amqp://%s",*f_amqpurl)
		var amqpServer *amqplistener.AMQPServer
		amqpServer = amqplistener.NewAMQPServer(amqpurl, true, *f_count, notifier)
		for {
				data := <-amqpServer.GetNotifier()
				//fmt.Printf("%v",data)
				c := cacheutil.ParseCollectdJSON(data)
				cacheserver.Put(*c)
		}
	}


}
