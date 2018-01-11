package MetricCache

import (
    "sync"

)
//NewCache ...
func NewCache() *Cache {
  return &Cache{
    hosts: make(map[string]*Plugins),
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
//NewPlugins ...
func NewPlugins() *Plugins {
  return &Plugins{
    plugins: make(map[string]*Plugin),
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


//Cache ,map[host1]["Metrics name"]=json string for now
type Cache struct {
    hosts map[string]*Plugins
    lock  *sync.RWMutex
}
/*Plugins ,sub types with plugin name fo host has values
Plugins,map[plugin_name] values are pointer to plugin*/
type Plugins struct {
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




// Put item with value v and key k into the hashtable
//ket is hostname and values is Metrics if not present add ?
//do we need to remove it?
//key is nod1.redhat.com and values is Metrics
// check if that Metricrcs object exists
// need to have seprate ut for Metrics so we don't lock entire hash of host
func (ht *Cache) Put(hostname string) {
    ht.lock.Lock()
    defer ht.lock.Unlock()
    //if the host aready exists it overrides... and deletes all child?
    //better to handle in application checkif key exists?
    if ht.hosts == nil { //assume first time it is alwasy nil so don'tdo in else
      ht.hosts = make(map[string]*Plugins)
    }
    ht.hosts[hostname] = NewPlugins()
  }


// Put item with value v and key k into the hashtable
// key is metric name and values are of type Metric
func (ht *Plugins) Put(pluginname string, plugin Plugin) {
    ht.lock.Lock()
    defer ht.lock.Unlock()
    if ht.plugins == nil {
        ht.plugins = make(map[string]*Plugin)
    }
    ht.plugins[pluginname]=NewPlugin()
    ht.plugins[pluginname].Put(plugin)
}

// mutable immutable .. try to handle it
func (ht *Plugin) Put(plugin Plugin)  {
  ht.lock.Lock()
  defer ht.lock.Unlock()
  //newPlugin:=Plugin{}
  ht.metrictype=plugin.metrictype
  ht.name=plugin.name
  ht.desc=plugin.desc
  ht.labels=NewLabel()
  ht.datasource=NewDataSource()
  for key,value :=range plugin.labels.items{
    ht.labels.items[key]=value
  }
  for key,value :=range plugin.datasource.ds{
    ht.datasource.ds[key]=value
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
func (ht *DataSource) Put(dsname string, dsvalue float64) {
    //ht.lock.Lock()
    //defer ht.lock.Unlock()
    if ht.ds == nil {
        ht.ds = make(map[string]float64)
    }
    ht.ds[dsname] = dsvalue
}


// Remove item with key k from hashtable
func (ht *Cache) Remove(hostname string) {
    ht.lock.Lock()
    defer ht.lock.Unlock()
    delete(ht.hosts, hostname)
}

// Get item with key k from the hashtable
func (ht *Cache) Get(hostname string) *Plugins {
    ht.lock.RLock()
    defer ht.lock.RUnlock()
    return ht.hosts[hostname]
}

// Get item with key k from the hashtable
func (ht *Label) Get(labelname string) string {
    //ht.lock.RLock()
    //defer ht.lock.RUnlock()
    return ht.items[labelname]
}
// Get item with key k from the hashtable
func (ht *DataSource) Get(dsname string) float64 {
    ht.lock.RLock()
    defer ht.lock.RUnlock()
    return ht.ds[dsname]
}

// Get item with key k from the hashtable
func (ht *Plugins) Get(pluginname string) *Plugin {
    ht.lock.RLock()
    defer ht.lock.RUnlock()
    return ht.plugins[pluginname]
}

// Size returns the number of the hashtable elements
func (ht *Cache) Size() int {
    ht.lock.RLock()
    defer ht.lock.RUnlock()
    return len(ht.hosts)
}
func (ht *Plugins) Size() int {
    ht.lock.RLock()
    defer ht.lock.RUnlock()
    return len(ht.plugins)
}
