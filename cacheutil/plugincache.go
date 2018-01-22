package cacheutil

import (
	"sync"
)

//Cache   ...cache by hostname
type Cache map[string]*ShardedPluginCache

/*ShardedPluginCache  ,sub types with plugin name fo host has values
ShardedPluginCache,   map[plugin_name] values are pointer to plugin*/
type ShardedPluginCache struct {
	Plugins map[string]*Plugin
	lock    *sync.RWMutex
}

//Label  ...
type Label struct {
	Items map[string]string
	lock  *sync.RWMutex
}

//DataSource ... data source name and value
type DataSource struct {
	Ds   map[string]float64
	lock *sync.RWMutex
}

//Plugin ...
type Plugin struct { // Size returns the number of the metric elements
	Metrictype string
	Name       string
	Desc       string
	Labels     *Label
	Datasource *DataSource
	lock       *sync.RWMutex
}

//NewCache   ...
func NewCache() Cache {
	return make(Cache)
}

//NewShardedPluginCache  . create new  sharded BufferCache
func NewShardedPluginCache() *ShardedPluginCache {
	return &ShardedPluginCache{
		Plugins: make(map[string]*Plugin),
		lock:    new(sync.RWMutex),
	}
}

//NewLabel ...
func NewLabel() *Label {
	return &Label{
		Items: make(map[string]string),
		lock:  new(sync.RWMutex),
	}
}

//NewDataSource  ...  Creates new datasource as pointer
func NewDataSource() *DataSource {
	return &DataSource{
		Ds:   make(map[string]float64),
		lock: new(sync.RWMutex),
	}
}

//NewPlugin  ...
func NewPlugin() *Plugin {
	return &Plugin{
		Labels:     NewLabel(),
		Datasource: NewDataSource(),
		lock:       new(sync.RWMutex),
	}
}

//Put   ..
func (c Cache) Put(hostname string) (shard *ShardedPluginCache) {
	return c.SetShard(hostname)
}

//Get ...
func (c *Cache) Get(hostname string) (shard *ShardedPluginCache) {
	return c.GetShard(hostname)
}

//SetShard   ...
func (c Cache) SetShard(hostname string) (shard *ShardedPluginCache) {
	shard = c[hostname]
	if shard == nil {
		c[hostname] = NewShardedPluginCache()
		shard = c[hostname]
	}

	return shard
}

//GetShard .... add shard
func (c Cache) GetShard(hostname string) (shard *ShardedPluginCache) {
	shard = c[hostname]
	if shard == nil {
		return c.SetShard(hostname)
	}
	return shard
}

// Put item with value v and key k into the hashtable
// key is metric name and values are of type Metric
func (shard *ShardedPluginCache) Put(pluginname string, plugin Plugin) {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	if shard.Plugins == nil {
		shard.Plugins = make(map[string]*Plugin)
	}
	if shard.Plugins[pluginname] == nil {
		shard.Plugins[pluginname] = NewPlugin()
	}
	shard.Plugins[pluginname].Put(plugin)

}
//GetPluginByName   ..
func (shard *ShardedPluginCache) GetPluginByName(pluginname string) *Plugin {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	if shard.Plugins == nil {
		shard.Plugins = make(map[string]*Plugin)
	}
	if shard.Plugins[pluginname] == nil {
		shard.Plugins[pluginname] = NewPlugin()
	}
	return shard.Plugins[pluginname]

}
//UpdateLabel  ...
func (p *Plugin) UpdateLabel(labels map[string]string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	for key, value := range labels {
		p.Labels.Items[key] = value
	}
}
//UpdateDataSource   ....
func (p *Plugin) UpdateDataSource(datasource map[string]float64) {
	p.lock.Lock()
	defer p.lock.Unlock()
	for key, value := range datasource {
		p.Datasource.Ds[key] = value
	}
}

//Add ...
func (p *Plugin) Add(metrictype string, pluginname string, desc string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	//newPlugin:=Plugin{}
	p.Metrictype = metrictype
	p.Name = pluginname
	p.Desc = desc
	if p.Labels == nil {
		p.Labels = NewLabel()
	}
	if p.Datasource == nil {
		p.Datasource = NewDataSource()
	}

}

//Put  mutable immutable .. try to handle it
func (p *Plugin) Put(plugin Plugin) {
	p.lock.Lock()
	defer p.lock.Unlock()
	//newPlugin:=Plugin{}
	p.Metrictype = plugin.Metrictype
	p.Name = plugin.Name
	p.Desc = plugin.Desc
	p.Labels = NewLabel()
	p.Datasource = NewDataSource()
	for key, value := range plugin.Labels.Items {
		p.Labels.Items[key] = value
	}
	for key, value := range plugin.Datasource.Ds {
		p.Datasource.Ds[key] = value
	}

}

// Put item with value v and key k into the hashtable
func (l *Label) Put(labelname string, labelvalue string) {
	//ht.lock.Lock()
	//defer ht.lock.Unlock()
	if l.Items == nil {
		l.Items = make(map[string]string)
	}
	l.Items[labelname] = labelvalue
}

//Put  .ShardedPluginCache
func (d *DataSource) Put(dsname string, dsvalue float64) {
	//ht.lock.Lock()
	//defer ht.lock.Unlock()
	if d.Ds == nil {
		d.Ds = make(map[string]float64)
	}
	d.Ds[dsname] = dsvalue
}

// Remove item with key k from hashtable
//Remove remove plugin for a given host
func (c Cache) Remove(hostname string, pluginname string) {
	shard := c.GetShard(hostname)
	shard.RemovePlugin(pluginname)
}

//RemovePlugin   remove at host level
func (shard ShardedPluginCache) RemovePlugin(pluginname string) {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	delete(shard.Plugins, pluginname)
}

//Removehost    .
func (c Cache) Removehost(hostname string) {
	delete(c, hostname)
}

// Remove item with key k from hashtable
func (l *Label) Remove(name string) {
	l.lock.Lock()
	defer l.lock.Unlock()
	delete(l.Items, name)
}

// Get item with key k from the hashtable
func (l *Label) Get(labelname string) string {
	//ht.lock.RLock()
	//defer ht.lock.RUnlock()
	return l.Items[labelname]
}

// Get item with key k from the hashtable
func (d *DataSource) Get(dsname string) float64 {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.Ds[dsname]
}

// Get item with key k from the hashtable
func (shard *ShardedPluginCache) Get(pluginname string) *Plugin {
	shard.lock.RLock()
	defer shard.lock.RUnlock()
	return shard.Plugins[pluginname]
}

// Size returns the number of the hashtable elements
func (c Cache) Size() int {
	return len(c)
}

// Size  .
func (shard *ShardedPluginCache) Size() int {
	shard.lock.RLock()
	defer shard.lock.RUnlock()
	return len(shard.Plugins)
}
