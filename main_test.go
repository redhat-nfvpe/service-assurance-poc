package main

import (
	"fmt"
	"github.com/aneeshkp/service-assurance-goclient/cacheutil"

	"sync"
	"testing"
	"time"
)

func TestPut(t *testing.T) {
	var cacheserver = cacheutil.NewCacheServer()
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
				go cacheutil.GenrateSampleData(hostname, noofpluginperhosts, jsondata, cacheserver)
			}(hosts)

		}
		/*for _,shard :=range cacheserver.cache.hosts{
		  fmt.Printf("Whole map %d",len(shard.plugin))
		  }*/
		hostwaitgroup.Wait()
		time.Sleep(time.Second * 1)

		if size := cacheserver.GetCache().Size(); size != noofhosts {
			t.Errorf("wrong count of hosts, expected %d and got %d", noofhosts, size)
		}
		for hostname, plugins := range cacheserver.GetCache().GetHosts() {
			if size := plugins.Size(); size != 100 {
				t.Errorf("wrong count of plugin per host %s, expected 100 and got %d", hostname, size)
			}
		}

	}
	//after everything is done
	if size := cacheserver.GetCache().Size(); size != noofhosts {
		t.Errorf("wrong count of hosts, expected %d and got %d", noofhosts, size)
	}
}
