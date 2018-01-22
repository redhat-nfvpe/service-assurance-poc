package cacheutil

import (
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"sync"
)

var freeList = make(chan *inputData, 100)

/****************************************/
//inputData   this is inutdata send to cache server
type inputData struct {
	collectd Collectd
}

//InputDataV2 ...cache server converts it into this
type InputDataV2 struct {
	hosts map[string]*ShardedInputDataV2
	lock  *sync.RWMutex
}


//ShardedInputDataV2 ...
type ShardedInputDataV2 struct {
	plugin map[string]*Collectd
	lock   *sync.RWMutex
}
//NewInputDataV2 ...
func NewInputDataV2() InputDataV2 {
	return InputDataV2{
		hosts: make(map[string]*ShardedInputDataV2),
		lock:  new(sync.RWMutex),
	}

}
//NewShardedInputDataV2 ...
func NewShardedInputDataV2() *ShardedInputDataV2 {
	return &ShardedInputDataV2{
		plugin: make(map[string]*Collectd),
		lock:   new(sync.RWMutex),
	}
}
//Put   ..
func (i InputDataV2) Put(hostname string) {
	//mutex.Lock()
	i.lock.Lock()
	defer i.lock.Unlock()
	i.hosts[hostname] = NewShardedInputDataV2()
	//i.hosts[hostname] = nil
	//mutex.UnLock()
}

//GetHosts  Get All hosts
func (i InputDataV2) GetHosts() map[string]*ShardedInputDataV2 {
	//mutex.Lock()
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.hosts
	//i.hosts[hostname] = nil
	//mutex.UnLock()
}

//GetShard  ..
func (i InputDataV2) GetShard(hostname string) *ShardedInputDataV2 {
	//GetShard .... add shard
	//i.lock.Lock()
	if i.hosts[hostname] == nil {
		i.Put(hostname)
	}

	return i.hosts[hostname]

}

//GetCollectD   ..
func (shard *ShardedInputDataV2) GetCollectD(pluginname string) Collectd {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	return *shard.plugin[pluginname]
}

//Size no of plugin per shard
func (i InputDataV2) Size() int {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return len(i.hosts)

}

//Size no of plugin per shard
func (shard *ShardedInputDataV2) Size() int {
	shard.lock.RLock()
	defer shard.lock.RUnlock()
	return len(shard.plugin)

}

//SetCollectD ...
func (shard *ShardedInputDataV2) SetCollectD(collectd Collectd) {
	shard.lock.Lock()
	defer shard.lock.Unlock()

	if shard.plugin[collectd.Plugin] == nil {
		shard.plugin[collectd.Plugin] = &Collectd{}
		shard.plugin[collectd.Plugin].Values = collectd.Values
		shard.plugin[collectd.Plugin].Dstypes = collectd.Dstypes
		shard.plugin[collectd.Plugin].Dsnames = collectd.Dsnames
		shard.plugin[collectd.Plugin].Time = collectd.Time
		shard.plugin[collectd.Plugin].Interval = collectd.Interval
		shard.plugin[collectd.Plugin].Host = collectd.Host
		shard.plugin[collectd.Plugin].Plugin = collectd.Plugin
		shard.plugin[collectd.Plugin].PluginInstance = collectd.PluginInstance
		shard.plugin[collectd.Plugin].Type = collectd.Type
		shard.plugin[collectd.Plugin].TypeInstance = collectd.TypeInstance
		shard.plugin[collectd.Plugin].SetNew(true)
	} else {
		shard.plugin[collectd.Plugin].Values = collectd.Values
		shard.plugin[collectd.Plugin].Dsnames = collectd.Dsnames
		shard.plugin[collectd.Plugin].Dstypes = collectd.Dstypes
		shard.plugin[collectd.Plugin].Time = collectd.Time
		if shard.plugin[collectd.Plugin].PluginInstance != collectd.PluginInstance {
			shard.plugin[collectd.Plugin].PluginInstance = collectd.PluginInstance
		}
		if shard.plugin[collectd.Plugin].Type != collectd.Type {
			shard.plugin[collectd.Plugin].Type = collectd.Type
		}
		if shard.plugin[collectd.Plugin].TypeInstance != collectd.TypeInstance {
			shard.plugin[collectd.Plugin].TypeInstance = collectd.TypeInstance
		}
		shard.plugin[collectd.Plugin].SetNew(true)
	}
	//log.Printf("sharded  %v\n",shard.plugin[collectd.Plugin])

}

//CacheServer   ..
type CacheServer struct {
	cache InputDataV2
	ch    chan *inputData
}

//GetCache  Get All hosts
func (cs *CacheServer) GetCache() *InputDataV2 {
	return &cs.cache

}

//NewCacheServer   ...
func NewCacheServer() *CacheServer {

	server := &CacheServer{
		cache: NewInputDataV2(),
		ch:    make(chan *inputData),
	}
	// Spawn off the server's main loop immediately
	go server.loop()
	return server
}

//Put  ..
func (cs *CacheServer) Put(collectd Collectd) {
	cs.ch <- &inputData{collectd: collectd}
}

//GetNewMetric   generate Prometheus metrics
func (shard *ShardedInputDataV2) GetNewMetric(ch chan<- prometheus.Metric) {
	log.Println("****************************   PROMETHEUS IS HERE ********************")
	shard.lock.Lock()
	defer shard.lock.Unlock()
	for _, collectd := range shard.plugin {
		//1==1 ||
		if collectd.ISNew() {
			collectd.SetNew(false)
			for index := range collectd.Values {
				//fmt.Printf("Before new metric %v\n", collectd)
				m, err := NewMetric(*collectd, index)
				log.Printf("Generated new Metric: %#v\n", m)
				if err != nil {

					log.Printf("newMtric: %v\n", err)
					continue
				}

				ch <- m
			}
		} else {
			log.Println("Skipping old Meteric")
		}
	}
}
//loop   .
func (cs *CacheServer) loop() {
	// The built-in "range" clause can iterate over channels,
	// amongst other things
	for {
		data := <-cs.ch
		shard := cs.cache.GetShard(data.collectd.Host)
		shard.SetCollectD(data.collectd)
		// Reuse buffer if there's room.
		select {
		case freeList <- data:
			// Buffer on free list; nothing more to do.
		default:
			// Free list full, just carry on.
		}
	}


}
