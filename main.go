package main

import (
	"fmt"
	"github.com/aneeshkp/service-assurance-goclient/cacheutil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"math/rand"
	"net/http"
	"time"
)

var (
	lastPush = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "collectd_last_push_timestamp_seconds",
			Help: "Unix timestamp of the last received collectd metrics push in seconds.",
		},
	)
)

//meterics  ... can I send cacheutil
/*func meterics(w http.ResponseWriter, r *http.Request, cache * cacheutil.Cache) {
	fmt.Fprintf(w, "I got somes metrics for you.. do you like it")
}*/
/*

[
   {
     "values":  [1901474177],
     "dstypes":  ["counter"],
     "dsnames":    ["value"],
     "time":      1280959128,
     "interval":          10,
     "host":            "leeloo.octo.it",
     "plugin":          "cpu",
     "plugin_instance": "0",
     "type":            "cpu",
     "type_instance":   "idle"
   }
 ]*/

//[{"values":[1],"dstypes":["gauge"],"dsnames":["value"],
//"time":1516043586.976,"interval":0.005,"host":"trex","plugin":"sysevent",
//"plugin_instance":"","type":"gauge","type_instance":"",
//"meta":{"@timestamp":"2018-01-15T19:13:06.971065+00:00","@source_host":"trex",
//"@message":"Jan 15 19:13:06 systemd:Starting Dynamic System Tuning Daemon...","facility":"daemon",
//"severity":"info","program":"systemd","processid":"-"}}]

//generateCollectdJson   for samples
func generateCollectdJson(hostname string, pluginname string) string {
	return `{
      "values":  [0.0,0.0],
      "dstypes":  ["gauge","guage"],
      "dsnames":    ["value1","value2"],
      "time":      0,
      "interval":          10,
      "host":            "hostname",
      "plugin":          "pluginname",
      "plugin_instance": "0",
      "type":            "pluginname",
      "type_instance":   "idle"
    }`
}

/****************************************/
type inputData struct {
	host       string
	pluginname string
	collectd   cacheutil.Collectd
}

type CacheServer struct {
	cache cacheutil.PrometehusCollector
	ch    chan inputData
}

func NewCacheServer() *CacheServer {

	server := &CacheServer{
		// make() creates builtins like channels, maps, and slices
		cache: cacheutil.NewPrometehusCollector(),
		ch:    make(chan inputData),
	}
	// Spawn off the server's main loop immediately
	go server.loop()
	return server
}

func (s *CacheServer) Put(hostname string, pluginname string, collectd cacheutil.Collectd) {
	fmt.Println("Putting data")
	s.ch <- inputData{host: hostname, pluginname: pluginname, collectd: collectd}
}
func (s *CacheServer) loop() {
	// The built-in "range" clause can iterate over channels,
	// amongst other things
	for {
		select {
		case data := <-s.ch:
			fmt.Printf("got message %v", data)
			// convert this data to prometheus cache
			shard := s.cache.GetShard(data.host)
			for i := range data.collectd.Values {
				m, err := cacheutil.NewMetric(data.collectd, i)
				if err != nil {
					log.Errorf("newMetric: %v", err)
					continue
				}
				metric_name := cacheutil.NewName(data.collectd, i)
				shard.Put(metric_name, m)
			}

			//build the cache
			//default:
			//  fmt.Println("Nothing to print")
		}
	}

	// Handle the command

}

/*************** HTTP HANDLER***********************/
type cacheHandler struct {
	cache *cacheutil.PrometehusCollector
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
	ch <- lastPush.Desc()
}

// Collect implements prometheus.Collector.
func (c *cacheHandler) Collect(ch chan<- prometheus.Metric) {
	//ch <- lastPush
	for hostname, m := range *c.cache {
		//fmt.Fprintln(w, hostname)
		for _, m := range m.GetMetrics(hostname) {
			//fmt.Fprintln(w, k.)
			ch <- *m
		}
	}
}

func main() {
	//I just learned this  from here http://www.alexedwards.net/blog/a-recap-of-request-handling
	/*
	   Processing HTTP requests with Go is primarily abo*testing.T)ut two things: ServeMuxes and Handlers.
	   The http.ServeMux is itself an http.Handler, so it can be passed into http.ListenAndServe.
	*/

	//var caches=make(cacheutil.Cache)
	var cacheserver = NewCacheServer()
	//nodeExport :=for i:=0;i<100;i++ { http.NewServeMux()
	myHandler := &cacheHandler{cache: &cacheserver.cache}
	/*	s := &http.Server{
	    Addr:           ":9002",
	    Handler:      myHandler  ,
	}*/

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

	//prometheus.MustRegister(myHandler)
	//s.Handle(*metricsPath, prometheus.Handler())

	//nodeExport.HandleFunc("/meterics", meterics(&cache))
	/***** use channel to pass the variable?????
	  channel is blocking... have to find a better way.. may be send pointer to cache
	*/

	// send it to its own g rountine
	//newbie I must be making ton of mistakes :-(
	//don't know how to handle if this goes down... do we need to restart whole app?
	/// need to do self rstarting thing

	//  populateCacheWithHosts(100,"redhat.bosoton.nfv",&caches)
	go func() {
		//http.ListenAndServe()
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	for {
		//sleep for 2 secs
		/*for hostname,pluginCache:= range caches{
		        setPlugin(hostname,pluginCache )
		}*/
		for i := 0; i < 1000; i++ {
			//100o hosts
			for j := 0; j < 100; j++ {
				//100 plugins

				var incoming_json = generateCollectdJson("hostname", "pluginname")
				var c = cacheutil.Collectd{}
				cacheutil.ParseCollectdJson(&c, incoming_json)

				// i have struct now filled with json data
				//convert this to prometheus format????
				var hostname = fmt.Sprintf("%s_%d", "redhat.bosoton.nfv", i)
				var pluginname = fmt.Sprintf("%s_%d", "plugin_name", j)
				c.Host = hostname
				c.Plugin = pluginname
				//to do I need to implment my own unmarshaller for this to work
				c.Dstypes[0] = "gauge"
				c.Dstypes[1] = "gauge"

				c.Dsnames[0] = "value1"
				c.Dsnames[1] = "value2"

				c.Values[0] = rand.Float64()
				c.Values[1] = rand.Float64()

				c.Time = (time.Now().UnixNano()) / 1000000
				fmt.Printf("incoming json %s\n", incoming_json)
				fmt.Printf("%v\n", c)

				cacheserver.Put(hostname, pluginname, c)
			}
		}
		time.Sleep(time.Second * 1)
	}

}
