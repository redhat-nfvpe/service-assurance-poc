package main

import (
	"fmt"
	"github.com/aneeshkp/service-assurance-goclient/cacheutil"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestPut(t *testing.T) {
	var cacheserver = NewCacheServer()
	var noofiteration, noofhosts, noofpluginperhosts int = 4, 2, 100

	var hostwaitgroup sync.WaitGroup

	var jsondata = cacheutil.GenerateCollectdJson("hostname", "pluginname")
	for times := 1; times <= noofiteration; times++ {
		hostwaitgroup.Add(noofhosts)
		for hosts := 0; hosts < noofhosts; hosts++ {
			go func(host_id int) {
				defer hostwaitgroup.Done()
				//100o hosts
				//pluginChannel := make(chan cacheutil.Collectd)
				//for each host make it on go routine
				var hostname = fmt.Sprintf("%s_%d", "redhat.bosoton.nfv", host_id)
				//fmt.Printf("Iteration %d hostname %s\n",times,hostname)
				go gentestdata(hostname, noofpluginperhosts, jsondata, cacheserver)
			}(hosts)

		}
		/*for _,shard :=range cacheserver.cache.hosts{
		  fmt.Printf("Whole map %d",len(shard.plugin))
		  }*/
		hostwaitgroup.Wait()
		time.Sleep(time.Second * 1)

		if size := cacheserver.cache.Size(); size != noofhosts {
			t.Errorf("wrong count of hosts, expected %d and got %d", noofhosts, size)
		}
		for hostname, plugins := range cacheserver.cache.hosts {
			if size := plugins.Size(); size != 100 {
				t.Errorf("wrong count of plugin per host %s, expected 100 and got %d", hostname, size)
			}
		}

	}
	//after everything is done
	if size := cacheserver.cache.Size(); size != noofhosts {
		t.Errorf("wrong count of hosts, expected %d and got %d", noofhosts, size)
	}
}

func gentestdata(hostname string, plugincount int, collectdjson string, cacheserver *CacheServer) {
	//100 plugins
	for j := 0; j < plugincount; j++ {
		var pluginname = fmt.Sprintf("%s_%d", "plugin_name", j)
		//fmt.Printf("index value is ****%d\n",j)
		//fmt.Printf("Plugin_name%s\n",pluginname)
		go func() {
			c := cacheutil.ParseCollectdJSON(collectdjson)
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
			//fmt.Printf("incoming json %s\n", collectdjson)
			//fmt.Printf("Before putting %v\n", c)
			cacheserver.Put(*c)
		}()

	}
}
