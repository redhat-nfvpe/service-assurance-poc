package cacheutil
import (
  "sync"
)

//BufferCache    ...
type BufferCache map[string]*BufferCacheShard

//BufferCacheShard  Shard cache by host ,will help if operations are done by hostname
//also locking is reduced to sub sets
type BufferCacheShard struct {
    data map[string]string
    lock  *sync.RWMutex
}


//NewBufferCache   ...
func NewBufferCache() BufferCache {
  return make(BufferCache)
}
//NewBufferCacheShard   create new  sharded BufferCache
func NewBufferCacheShard() *BufferCacheShard {
  return &BufferCacheShard{
      data: make(map[string]string),
      lock: new(sync.RWMutex),
    }
}

//Get ... get sharded cache by hostname
func (bc BufferCache) Get(hostname string) *BufferCacheShard {
  shard := bc.GetShard(hostname)
  shard.lock.RLock()
  defer shard.lock.RUnlock()
  return shard
}
//GetShard .... add shard
func (bc BufferCache) GetShard(hostname string) (shard *BufferCacheShard) {
     shard=bc[hostname]
     if shard == nil{
       bc[hostname]=NewBufferCacheShard()
       shard=bc[hostname]
    }
    return shard
}

//Set  .. set  plugin data at host level
func (bc BufferCache) Set(hostname string,pluginname string, data string) {
  shard := bc.GetShard(hostname)
  shard.lock.Lock()
  defer shard.lock.Unlock()
  shard.data[pluginname]=data
}


//Remove remove plugin for a given host
func (bc BufferCache) Remove(hostname string, pluginname string) {
  shard := bc.GetShard(hostname)
  shard.lock.Lock()
  defer shard.lock.Unlock()
  delete(shard.data, pluginname)
}
//Remove   remove at host level
func (bc BufferCacheShard) Remove(pluginname string) {
  bc.lock.Lock()
  defer bc.lock.Unlock()
  delete(bc.data, pluginname)
}
