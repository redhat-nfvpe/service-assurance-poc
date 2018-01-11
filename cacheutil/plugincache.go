package cacheutil

import (
    "sync"

)
//Cache   ...cache by hostname
type Cache map[string]*ShardedPluginCache


/*Plugins ,sub types with plugin name fo host has values
Plugins,map[plugin_name] values are pointer to plugin*/
type ShardedPluginCache struct {
  plugins map[string]*Plugin
  lock  *sync.RWMutex
}
//Label  ...
type Label struct{
  items map[string]string
  lock  *sync.RWMutex
}
//DataSource ... data source name and value
type DataSource struct{
  ds map[string]float64
  lock  *sync.RWMutex
}
//Plugin ...
type Plugin struct { // Size returns the number of the metric elements
  metrictype string
  name string
  desc string
  labels *Label
  datasource *DataSource
  lock  *sync.RWMutex
}

//NewBufferCache   ...
func NewCache() Cache {
  return make(Cache)
}
//NewBufferCacheShard   create new  sharded BufferCache
func NewShardedPluginCache() *ShardedPluginCache {
  return &ShardedPluginCache{
      plugins: make(map[string]*Plugin),
      lock: new(sync.RWMutex),
    }
}


//NewLabel ...
func NewLabel() *Label {
  return &Label{
    items: make(map[string]string),
    lock: new(sync.RWMutex),

  }
}
//NewDataSource  ...  Creates new datasource as pointer
func NewDataSource() *DataSource {
  return &DataSource{
    ds: make(map[string]float64),
    lock: new(sync.RWMutex),

  }
}

//NewPlugin  ...
func NewPlugin() *Plugin {
  return &Plugin{
    labels: NewLabel(),
    datasource: NewDataSource(),
    lock: new(sync.RWMutex),
  }
}


//SetShard
func (c Cache) Put(hostname string) (shard *ShardedPluginCache) {
    return c.SetShard(hostname)
}
func (c *Cache) Get(hostname string) (shard *ShardedPluginCache) {
    return c.GetShard(hostname)
}

func (c Cache) SetShard(hostname string) (shard *ShardedPluginCache) {
     shard=c[hostname]
     if shard == nil{
       c[hostname]=NewShardedPluginCache()
       shard=c[hostname]
    }

    return shard
}

//GetShard .... add shard
func (c Cache) GetShard(hostname string) (shard *ShardedPluginCache) {
     shard=c[hostname]
     if shard == nil{
       return c.SetShard(hostname)
    }
    return shard
}




// Put item with value v and key k into the hashtable
// key is metric name and values are of type Metric
func (shard *ShardedPluginCache) Put(pluginname string, plugin Plugin) {
   shard.lock.Lock()
   defer shard.lock.Unlock()
   if shard.plugins == nil {
        shard.plugins = make(map[string]*Plugin)
    }
    if shard.plugins[pluginname] ==nil{
     shard.plugins[pluginname]=NewPlugin()
    }
    shard.plugins[pluginname].Put(plugin)

}

//Put  mutable immutable .. try to handle it
func (p *Plugin) Put(plugin Plugin)  {
  p.lock.Lock()
  defer p.lock.Unlock()
  //newPlugin:=Plugin{}
  p.metrictype=plugin.metrictype
  p.name=plugin.name
  p.desc=plugin.desc
  p.labels=NewLabel()
  p.datasource=NewDataSource()
  for key,value :=range plugin.labels.items{
    p.labels.items[key]=value
  }
  for key,value :=range plugin.datasource.ds{
    p.datasource.ds[key]=value
}

}



// Put item with value v and key k into the hashtable
func (ht *Label) Put(labelname string, labelvalue string) {
    //ht.lock.Lock()
    //defer ht.lock.Unlock()
    if ht.items == nil {
        ht.items = make(map[string]string)
    }
    ht.items[labelname] = labelvalue
}

//Put  .ShardedPluginCache
func (ht *DataSource) Put(dsname string, dsvalue float64) {
    //ht.lock.Lock()
    //defer ht.lock.Unlock()
    if ht.ds == nil {
        ht.ds = make(map[string]float64)
    }
    ht.ds[dsname] = dsvalue
}


// Remove item with key k from hashtable
//Remove remove plugin for a given host
func (c Cache) Remove(hostname string, pluginname string) {
  shard := c.GetShard(hostname)
  shard.RemovePlugin(pluginname)
}
//Remove   remove at host level
func (shard ShardedPluginCache) RemovePlugin(pluginname string) {
  shard.lock.Lock()
  defer shard.lock.Unlock()
  delete(shard.plugins, pluginname)
}

func (c Cache) Removehost(hostname string) {
    delete(c, hostname)
}


// Remove item with key k from hashtable
func (l *Label) Remove(name string) {
    l.lock.Lock()
    defer l.lock.Unlock()
    delete(l.items, name)
}


// Get item with key k from the hashtable
func (l *Label) Get(labelname string) string {
    //ht.lock.RLock()
    //defer ht.lock.RUnlock()
    return l.items[labelname]
}
// Get item with key k from the hashtable
func (d *DataSource) Get(dsname string) float64 {
    d.lock.RLock()
    defer d.lock.RUnlock()
    return d.ds[dsname]
}

// Get item with key k from the hashtable
func (shard *ShardedPluginCache) Get(pluginname string) *Plugin {
    shard.lock.RLock()
    defer shard.lock.RUnlock()
    return shard.plugins[pluginname]
}

// Size returns the number of the hashtable elements
func (c Cache) Size() int {
    return len(c)
}
// Size  .
func (shard *ShardedPluginCache) Size() int {
    shard.lock.RLock()
    defer shard.lock.RUnlock()
    return len(shard.plugins)
}
