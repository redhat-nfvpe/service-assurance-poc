package cacheutil

import (
	"github.com/prometheus/client_golang/prometheus"
  "github.com/aneeshkp/service-assurance-goclient/incoming"
	"log"
	"sync"
	"errors"
)

var freeList = make(chan *Incoming, 100)

/****************************************/
//Incoming  this is inut data send to cache server
//Incoming  ..its of type collectd or anything else
type Incoming struct {
	data interface{}
}

//IncomingCache cache server converts it into this
type IncomingCache struct {
	hosts map[string]*ShardedIncomingCache
	lock  *sync.RWMutex
}

//types of sharded cache collectd, influxdb etc
//type InputDataV2 map[string]*ShardedInputDataV2
//ShardedIncomingCache  ..
type ShardedIncomingCache struct {
	plugin map[string]*incoming.Interface
	lock   *sync.RWMutex
}

//IncomingCache   .. .
func NewCache() IncomingCache {
	return IncomingCache{
		hosts: make(map[string]*ShardedIncomingCache),
		lock:  new(sync.RWMutex),
	}
}

//NewShardedIncomingCache   .
func NewShardedIncomingCache() *ShardedIncomingCache {
	return &ShardedIncomingCache{
		plugin: make(map[string]*incoming.Interface),
		lock:   new(sync.RWMutex),
	}
}

//PUT   ..
func (i IncomingCache) Put(key string) {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.hosts[key] = NewShardedIncomingCache()
}

//GetHosts  Get All hosts
func (i IncomingCache) GetHosts() map[string]*ShardedIncomingCache {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.hosts
}

//GetShard  ..
func (i IncomingCache) GetShard(key string) *ShardedIncomingCache {
	//GetShard .... add shard
	//i.lock.Lock()
	if i.hosts[key] == nil {
		i.Put(key)
	}

	return i.hosts[key]

}

//GetCollectD   ..
func (shard *ShardedIncomingCache) GetData(pluginname string) incoming.Interface {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	return shard.plugin[pluginname]
}

//Size no of plugin per shard
func (i IncomingCache) Size() int {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return len(i.hosts)

}

//Size no of plugin per shard
func (shard *ShardedIncomingCache) Size() int {
	shard.lock.RLock()
	defer shard.lock.RUnlock()
	return len(shard.plugin)

}

//SetData
func (shard *ShardedIncomingCache) SetData(data interface{}) error {
	shard.lock.Lock()
	defer shard.lock.Unlock()
  if collectd, ok := data.(incoming.Collectd); ok {
		shard.lock.Lock()
		defer shard.lock.Unlock()
		if shard.plugin[collectd.GetName()] == nil {
				shard.plugin[collectd.GetName()] =*incoming.CreateNewCollectd()
		}
		 shard.plugin[collectd.GetName()].SetData(data)
		 return nil
	}else{
    return errors.New("unknow data type while setting data")
	}


}

//CacheServer   ..
type CacheServer struct {
	cache IncomingCache
	ch    chan Incoming
}

//GetCache  Get All hosts
func (c *CacheServer) GetCache() *IncomingCache {
	return &c.cache
}

//NewCacheServer   ...
func NewCacheServer(cacheType incoming.IncomingDataType) *CacheServer {
	server := &CacheServer{
		cache: NewCache(),
		ch:    make(chan Incoming),
	}
	// Spawn off the server's main loop immediately
	go server.loop()
	return server
}

func (s *CacheServer) Put(data interface{}) {
	if collectd, ok := data.(incoming.Collectd); ok {
		s.ch <- Incoming{data: collectd}
	}

}

//GetNewMetric   generate Prometheus metrics
func (shard *ShardedIncomingCache) GetNewMetric(ch chan<- prometheus.Metric) {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	for _, incomingInterface := range shard.plugin {
		if collectd, ok := incomingInterface.(incoming.Collectd); ok {
		if collectd.ISNew() {
			collectd.SetNew(false)
			for index := range collectd.Values {
				//fmt.Printf("Before new metric %v\n", collectd)
				m, err := NewMetric(*collectd, index)
				if err != nil {
					log.Printf("newMetric: %v", err)
					continue
				}

				ch <- m
			}
		}
		}
	}
}
func (s *CacheServer) loop() {
	// The built-in "range" clause can iterate over channels,
	// amongst other things
	for {
		data := <-s.ch
		shard := s.cache.GetShard(data.collectd.Host)
		shard.SetCollectD(data.collectd)
		// Reuse buffer if there's room.
		select {
		case freeList <- data:
			// Buffer on free list; nothing more to do.
		default:
			// Free list full, just carry on.
		}
		/*select {
		case data := <-s.ch:
			//fmt.Printf("got message in channel %v", data)
			shard := s.cache.GetShard(data.collectd.Host)
			shard.SetCollectD(data.collectd)

		}*/
	}

}


//GenrateSampleData
func (cs *CacheServer)GenrateSampleData(key string, datacount int, jsonString string,datatype interface{}) {
	//100 plugins
	for j := 0; j < datacount; j++ {
		var pluginname = fmt.Sprintf("%s_%d", "plugin_name", j)
		go func() {
      switch datatype.(Type) {
      case incoming.Collectd:
        data=datatype.(incoming.Collectd).GenrateSampleData(key,pluginname,jsonstring)
        c.Host = hostname
  			c.Plugin = pluginname
  			c.Type = pluginname
  			c.Plugin_instance = pluginname
  			c.Dstypes[0] = "gauge"
  			c.Dstypes[1] = "gauge"
  			c.Dsnames[0] = "value1"
  			c.Dsnames[1] = "value2"
  			c.Values[0] = rand.Float64()
  			c.Values[1] = rand.Float64()
  			c.Time = float64((time.Now().UnixNano())) / 1000000

      }
			c := incoming.ParseInputJSON(json)
			c.Host = hostname
			c.Plugin = pluginname
			c.Type = pluginname
			c.Plugin_instance = pluginname
			c.Dstypes[0] = "gauge"
			c.Dstypes[1] = "gauge"
			c.Dsnames[0] = "value1"
			c.Dsnames[1] = "value2"
			c.Values[0] = rand.Float64()
			c.Values[1] = rand.Float64()
			c.Time = float64((time.Now().UnixNano())) / 1000000
			cs.Put(*c)
		}()
	}
}
