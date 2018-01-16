package main

import (
	"fmt"
	"github.com/aneeshkp/service-assurance-goclient/cacheutil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var (
	lastpulltime int64
	lastPull = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "collectd_last_pull_timestamp_seconds",
			Help: "Unix timestamp of the last received collectd metrics pull in seconds.",
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
//this is inutdata send to cache server
type inputData struct {
	collectd cacheutil.Collectd
}

//cache server converts it into this
type InputDataV2 struct {
	hosts map[string]*ShardedInputDataV2
	lock  *sync.RWMutex
}

//type InputDataV2 map[string]*ShardedInputDataV2

type ShardedInputDataV2 struct {
	plugin map[string]*cacheutil.Collectd
	lock   *sync.RWMutex
}

func NewInputDataV2() InputDataV2 {
	return InputDataV2{
		hosts: make(map[string]*ShardedInputDataV2),
		lock:  new(sync.RWMutex),
	}

}
func NewShardedInputDataV2() *ShardedInputDataV2 {
	return &ShardedInputDataV2{
		plugin: make(map[string]*cacheutil.Collectd),
		lock:   new(sync.RWMutex),
	}
}
func (i InputDataV2) Put(hostname string) {
	//mutex.Lock()
	i.lock.Lock()
	defer i.lock.Unlock()
	i.hosts[hostname] = NewShardedInputDataV2()
	//mutex.UnLock()
}
func (i InputDataV2) GetShard(hostname string) *ShardedInputDataV2 {
	//GetShard .... add shard
	var shard = i.hosts[hostname]
	fmt.Println(shard)
	if shard == nil {
		fmt.Println("Shard is empty")
		i.Put(hostname)
		fmt.Println(i.hosts[hostname])
	}
	return i.hosts[hostname]
}
func (shard *ShardedInputDataV2) GetCollectD(pluginname string) cacheutil.Collectd {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	return *shard.plugin[pluginname]
}

func (shard *ShardedInputDataV2) SetCollectD(collectd cacheutil.Collectd) {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	if shard.plugin[collectd.Plugin] == nil {
		shard.plugin[collectd.Plugin] = &cacheutil.Collectd{}
		shard.plugin[collectd.Plugin].Values = collectd.Values
		shard.plugin[collectd.Plugin].Dstypes = collectd.Dstypes
		shard.plugin[collectd.Plugin].Dsnames = collectd.Dsnames
		shard.plugin[collectd.Plugin].Time = collectd.Time
		shard.plugin[collectd.Plugin].Interval = collectd.Interval
		shard.plugin[collectd.Plugin].Host = collectd.Host
		shard.plugin[collectd.Plugin].Plugin = collectd.Plugin
		shard.plugin[collectd.Plugin].Plugin_instance = collectd.Plugin_instance
		shard.plugin[collectd.Plugin].Type = collectd.Type
		shard.plugin[collectd.Plugin].Type_instance = collectd.Type_instance
	} else {
		shard.plugin[collectd.Plugin].Values = collectd.Values
		shard.plugin[collectd.Plugin].Dsnames = collectd.Dsnames
		shard.plugin[collectd.Plugin].Dstypes = collectd.Dstypes
		shard.plugin[collectd.Plugin].Time = collectd.Time
		if shard.plugin[collectd.Plugin].Plugin_instance != collectd.Plugin_instance {
			shard.plugin[collectd.Plugin].Plugin_instance = collectd.Plugin_instance
		}
		if shard.plugin[collectd.Plugin].Type != collectd.Type {
			shard.plugin[collectd.Plugin].Type = collectd.Type
		}
		if shard.plugin[collectd.Plugin].Type_instance != collectd.Type_instance {
			shard.plugin[collectd.Plugin].Type_instance = collectd.Type_instance
		}

	}

}

type CacheServer struct {
	cache InputDataV2
	ch    chan inputData
}

func NewCacheServer() *CacheServer {

	server := &CacheServer{
		// make() creates builtins like channels, maps, and slices
		//cache: cacheutil.NewPrometehusCollector(),
		cache: NewInputDataV2(),
		ch:    make(chan inputData),
	}
	// Spawn off the server's main loop immediately
	go server.loop()
	return server
}

func (s *CacheServer) Put(collectd cacheutil.Collectd) {
	fmt.Println("Putting data")
	//s.ch <- inputData{host: hostname, pluginname: pluginname, collectd: collectd}
	s.ch <- inputData{collectd: collectd}

}
func (shard *ShardedInputDataV2) GetNewMetric(ch chan<- prometheus.Metric,lasttime int64) {
	for _, collectd := range shard.plugin {
		for i := range collectd.Values {
			if collectd.Time>lasttime{
				m, err := cacheutil.NewMetric(*collectd, i)
				if err != nil {
					log.Errorf("newMetric: %v", err)
					continue
				}
				ch <- m
			}
		}
	}
}
func (s *CacheServer) loop() {
	// The built-in "range" clause can iterate over channels,
	// amongst other things
	for {
		select {
		case data := <-s.ch:
			fmt.Printf("got message %v", data)

			shard := s.cache.GetShard(data.collectd.Host)
			fmt.Println("got shard")
			fmt.Printf("got shard%v", shard)
			shard.SetCollectD(data.collectd)
		}
	}

	// Handle the command

}

/*************** HTTP HANDLER***********************/
type cacheHandler struct {
	cache *InputDataV2
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
func (c *cacheHandler) Collect(ch chan<- prometheus.Metric) {
	lastPull.Set(float64(time.Now().UnixNano()) / 1e9)
	ch <- lastPull
	var previouspullTime=lastpulltime
	lastpulltime=int64(time.Now().UnixNano()) / 1e9
	for _, plugin := range c.cache.hosts {
		//fmt.Fprintln(w, hostname)
		plugin.GetNewMetric(ch,previouspullTime)
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
		for i := 0; i < 2; i++ {
			//100o hosts
			//pluginChannel := make(chan cacheutil.Collectd)
			var jsondata = generateCollectdJson("hostname", "pluginname")
			//for each host make it on go routine
			go func() {
				var hostname = fmt.Sprintf("%s_%d", "redhat.bosoton.nfv", i)
				gentestdata(hostname, 100, jsondata, cacheserver)
			   //collectdPluginData := <-pluginChannel
   			 //cacheserver.Put(collectdPluginData)
			}()

		}
		time.Sleep(time.Second * 1)
	}

}

func gentestdata(hostname string, plugincount int, collectdjson string, cacheserver *CacheServer) {
	//100 plugins
	for j := 0; j < plugincount; j++ {
		var pluginname = fmt.Sprintf("%s_%d", "plugin_name", j)
		fmt.Printf("index value is ****%d\n",j)
		fmt.Printf("Plugin_name%s\n",pluginname)
		var c = cacheutil.Collectd{}
		cacheutil.ParseCollectdJson(&c, collectdjson)
		// i have struct now filled with json data
		//convert this to prometheus format????

		c.Host = hostname
		c.Plugin = pluginname
		c.Type = pluginname
		c.Plugin_instance = pluginname
		//to do I need to implment my own unmarshaller for this to work
		c.Dstypes[0] = "gauge"
		c.Dstypes[1] = "gauge"
		c.Dsnames[0] = "value1"
		c.Dsnames[1] = "value2"
		c.Values[0] = rand.Float64()
		c.Values[1] = rand.Float64()
		c.Time = (time.Now().UnixNano()) / 1000000
		fmt.Printf("incoming json %s\n", collectdjson)
		fmt.Printf("%v\n", c)
		cacheserver.Put(c)
	}
}
