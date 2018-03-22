package cacheutil

import (
	"log"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redhat-nfvpe/service-assurance-poc/incoming"
	"github.com/redhat-nfvpe/service-assurance-poc/tsdb"

	//"errors"
	"fmt"
)

var freeList = make(chan *IncomingBuffer, 100)
var quitCacheServerCh = make(chan struct{})

//IncomingBuffer  this is inut data send to cache server
//IncomingBuffer  ..its of type collectd or anything else
type IncomingBuffer struct {
	data incoming.DataTypeInterface
}

//IncomingDataCache cache server converts it into this
type IncomingDataCache struct {
	hosts map[string]*ShardedIncomingDataCache
	lock  *sync.RWMutex
}

//ShardedIncomingDataCache types of sharded cache collectd, influxdb etc
//ShardedIncomingDataCache  ..
type ShardedIncomingDataCache struct {
	plugin map[string]incoming.DataTypeInterface
	lock   *sync.RWMutex
}

//NewCache   .. .
func NewCache() IncomingDataCache {
	return IncomingDataCache{
		hosts: make(map[string]*ShardedIncomingDataCache),
		lock:  new(sync.RWMutex),
	}
}

//NewShardedIncomingDataCache   .
func NewShardedIncomingDataCache() *ShardedIncomingDataCache {
	return &ShardedIncomingDataCache{
		plugin: make(map[string]incoming.DataTypeInterface),
		lock:   new(sync.RWMutex),
	}
}

//Put   ..
func (i IncomingDataCache) Put(key string) {
	i.lock.Lock()
	defer i.lock.Unlock()
	i.hosts[key] = NewShardedIncomingDataCache()
}

//GetHosts  Get All hosts
func (i IncomingDataCache) GetHosts() map[string]*ShardedIncomingDataCache {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.hosts
}

//GetShard  ..
func (i IncomingDataCache) GetShard(key string) *ShardedIncomingDataCache {
	//GetShard .... add shardGetCollectD
	//i.lock.Lock()
	if i.hosts[key] == nil {
		i.Put(key)
	}

	return i.hosts[key]

}

//GetData   ..
func (shard *ShardedIncomingDataCache) GetData(pluginname string) incoming.DataTypeInterface {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	return shard.plugin[pluginname]
}

//Size no of plugin per shard
func (i IncomingDataCache) Size() int {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return len(i.hosts)

}

//Size no of plugin per shard
func (shard *ShardedIncomingDataCache) Size() int {
	shard.lock.RLock()
	defer shard.lock.RUnlock()
	return len(shard.plugin)

}

//SetData  TODO : add generic
func (shard *ShardedIncomingDataCache) SetData(data incoming.DataTypeInterface) error {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	if shard.plugin[data.GetItemKey()] == nil {
		//TODO: change this to more generic later
		shard.plugin[data.GetItemKey()] = incoming.NewInComing(incoming.COLLECTD)
	}
	collectd := shard.plugin[data.GetItemKey()]
	collectd.SetData(data)
	return nil

	//return errors.New("unknow data type while setting data")

}

//CacheServer   ..
type CacheServer struct {
	cache IncomingDataCache
	ch    chan *IncomingBuffer
}

//GetCache  Get All hosts
func (cs *CacheServer) GetCache() *IncomingDataCache {
	return &cs.cache
}

//NewCacheServer   ...
func NewCacheServer() *CacheServer {
	server := &CacheServer{
		cache: NewCache(),
		ch:    make(chan *IncomingBuffer),
	}
	// Spawn off the server's main loop immediately
	go server.loop()
	return server
}

//Put   ..
func (cs *CacheServer) Put(incomingData incoming.DataTypeInterface) {
	var buffer *IncomingBuffer
	select {
	case buffer = <-freeList:
		//go one from buffer
	default:
		buffer = new(IncomingBuffer)
	}
	buffer.data = incomingData
	cs.ch <- buffer

}

//GetNewMetric   generate Prometheus metrics
func (shard *ShardedIncomingDataCache) GetNewMetric(ch chan<- prometheus.Metric) {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	for _, IncomingDataInterface := range shard.plugin {
		if collectd, ok := IncomingDataInterface.(*incoming.Collectd); ok {
			if collectd.ISNew() {
				collectd.SetNew(false)
				for index := range collectd.Values {
					m, err := tsdb.NewCollectdMetric(*collectd, index)
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
func (cs CacheServer) close() {
	<-quitCacheServerCh
	close(quitCacheServerCh)
}
func (cs CacheServer) loop() {
	// The built-in "range" clause can iterate over channels,
	// amongst other things
LOOP:
	for {
		// Reuse buffer if there's room.
		buffer := <-cs.ch
		shard := cs.cache.GetShard(buffer.data.GetKey())
		shard.SetData(buffer.data)
		select {
		case freeList <- buffer:
		// Buffer on free list; nothing more to do.
		case <-quitCacheServerCh:
			break LOOP
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

//GenrateSampleData  ....
func (cs *CacheServer) GenrateSampleData(key string, itemCount int, datatype incoming.DataTypeInterface) {
	//100 plugins
	for j := 0; j < itemCount; j++ {
		pluginname := fmt.Sprintf("%s_%d", "plugin_name_", j)
		//. defer wg.Done()
		newSample := datatype.GenerateSampleData(key, pluginname)
		cs.Put(newSample)

	}

}
