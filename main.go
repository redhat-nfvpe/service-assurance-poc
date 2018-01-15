package main

import (
	"fmt"
	"net/http"
  "time"
  "math/rand"
  "github.com/aneeshkp/service-assurance-goclient/cacheutil"
)
//meterics  ... can I send cacheutil
/*func meterics(w http.ResponseWriter, r *http.Request, cache * cacheutil.Cache) {
	fmt.Fprintf(w, "I got somes metrics for you.. do you like it")
}*/


/**********************************/
func populateCacheWithHosts(count int ,hostname string, cache *cacheutil.Cache)  {
  //hostDict:=Cache{}

  for i:=0;i<count;i++ {
    cache.Put(fmt.Sprintf("%s_%d", hostname,i))
  }

}
func getLabels(hostname string) cacheutil.Label{
  labels :=cacheutil.Label{}
  labels.Put("instance",hostname)
  //labels.Put("id",  strconv.Itoa(id))
  labels.Put("foo","bar")
  return labels
}
//get 100's of  metric for each host
func setPlugin(hostname string, pluginCache *cacheutil.ShardedPluginCache) {
  // initlaizepluginsDic:=
  //some common name
  pluginNames :=[]string{"interface","network","cpuutilization","memoryused","memoryfree"}
  // 100 plugin
  var plugins[100]string

  // generate 100 difference meteric names
  var j int
  for i:=0;i<20;i++ {
    for _,value:= range pluginNames{
    plugins[j]=fmt.Sprintf("%s_%s_%d", "metric",value,j)
    j++
    }
  }
  //data to types for all
  var data[2]string
  data[0]="rx"
  data[1]="tx"
  //for each host get 100 plugin

  for _, pluginNames:= range plugins{

    plugin:=cacheutil.NewPlugin()
    plugin.Metrictype ="guage"
    plugin.Name = pluginNames
    labels := getLabels(hostname)
    for key, value :=range labels.Items {
      plugin.Labels.Put(key,value)
    }
    for _, value := range data {
      plugin.Datasource.Put(value,rand.Float64())
    }
    //deference pointer befor sending
    pluginCache.Put(plugin.Name,*plugin)

    }
}

/****************************************/
type inputData struct{
	host string
	pluginname string
	jsondata string
}

type CacheServer struct {
    cache cacheutil.Cache
		ch chan inputData
}

func NewCacheServer() *CacheServer {

		server := &CacheServer{
		        // make() creates builtins like channels, maps, and slices
		       cache : cacheutil.NewCache(),
					 ch :make(chan inputData),
		    }
    // Spawn off the server's main loop immediately
    go server.loop()
    return server
}

func(s *CacheServer)put(hostname string,pluginname string, jsondata string){
	fmt.Println("Putting data")
	s.ch<-inputData{host:hostname,pluginname:pluginname,jsondata:jsondata}
}
func (s *CacheServer) loop() {
    // The built-in "range" clause can iterate over channels,
    // amongst other things
		for{
			select{
			case data:=<-s.ch:
				   fmt.Printf("got message %v",data)
					 shard:=s.cache.GetShard(data.host)

					 plugin:=shard.GetPluginByName(data.pluginname)
					 plugin.Add("gauge",data.pluginname,"description")

					 //build the cache
			//default:
				//  fmt.Println("Nothing to print")
			}
		}

        // Handle the command

}

type cacheHandler struct {
    cache *cacheutil.Cache
}

func (h *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
   for hostname,plugin:= range *h.cache {
    fmt.Fprintln(w, hostname)
		for k :=range plugin.Plugins{
			fmt.Fprintln(w, k)

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
	var cacheserver=NewCacheServer()
	//nodeExport := http.NewServeMux()
  myHandler:=&cacheHandler{cache: &cacheserver.cache}
	s := &http.Server{
        Addr:           ":9002",
        Handler:      myHandler  ,
    }


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
		s.ListenAndServe()
	}()

  for {
    //sleep for 2 secs
    /*for hostname,pluginCache:= range caches{
            setPlugin(hostname,pluginCache )
    }*/
		for i:=0;i<100;i++ {
			cacheserver.put(fmt.Sprintf("%s_%d", "redhat.bosoton.nfv",i),
										fmt.Sprintf("%s_%d", "plugin_name",i),"jsondata")

		}

		   time.Sleep(time.Second * 1)

  }


}
