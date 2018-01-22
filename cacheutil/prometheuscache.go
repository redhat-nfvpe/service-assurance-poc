package cacheutil

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	metricNameRe = regexp.MustCompile("[^a-zA-Z0-9_:]")
)

//PrometehusCollector   ..
type PrometehusCollector map[string]*ShardedPrometehusCollector

//ShardedPrometehusCollector ...  ,sub types with plugin name fo host has values
type ShardedPrometehusCollector struct {
	metric map[string]*prometheus.Metric
	lock   *sync.RWMutex
}

//NewPrometehusCollector   ...
func NewPrometehusCollector() PrometehusCollector {
	return make(PrometehusCollector)
}

//NewShardedPrometehusCollector  . create new  sharded BufferCache
func NewShardedPrometehusCollector() *ShardedPrometehusCollector {
	return &ShardedPrometehusCollector{
		metric: make(map[string]*prometheus.Metric),
		lock:   new(sync.RWMutex),
	}
}

//Put   ..
func (c PrometehusCollector) Put(hostname string) (shard *ShardedPrometehusCollector) {
	return c.SetShard(hostname)
}

//Get ...
func (c *PrometehusCollector) Get(hostname string) (shard *ShardedPrometehusCollector) {
	return c.GetShard(hostname)
}

//GetMetrics  ...
func (shard *ShardedPrometehusCollector) GetMetrics(hostname string) (metrics map[string]*prometheus.Metric) {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	return shard.metric
}

//SetShard   ...
func (c PrometehusCollector) SetShard(hostname string) (shard *ShardedPrometehusCollector) {
	shard = c[hostname]
	if shard == nil {
		c[hostname] = NewShardedPrometehusCollector()
		shard = c[hostname]
	}

	return shard
}

//GetShard .... add shard
func (c PrometehusCollector) GetShard(hostname string) (shard *ShardedPrometehusCollector) {
	shard = c[hostname]
	if shard == nil {
		return c.SetShard(hostname)
	}
	return shard
}

// Put item with value v and key k into the hashtable
// key is metric name and values are of type Metric
func (shard *ShardedPrometehusCollector) Put(pluginname string, metric prometheus.Metric) {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	if shard.metric == nil {
		shard.metric = make(map[string]*prometheus.Metric)
	}

	shard.metric[pluginname] = &metric

}

//Get  ...
func (shard ShardedPrometehusCollector) Get(pluginname string) prometheus.Metric {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	if shard.metric == nil {
		shard.metric = make(map[string]*prometheus.Metric)
	}

	return *shard.metric[pluginname]

}

// Size returns the number of the hashtable elements
func (c PrometehusCollector) Size() int {
	return len(c)
}

// Size  .
func (shard *ShardedPrometehusCollector) Size() int {
	shard.lock.RLock()
	defer shard.lock.RUnlock()
	return len(shard.metric)
}

//NewName  ....
func NewName(vl Collectd, index int) string {
	var name string
	if vl.Plugin == vl.Type {
		name = "serviceassurancecollectd_" + vl.Type
	} else {
		name = "serviceassurancecollectd_" + vl.Plugin + "_" + vl.Type
	}

	if dsname := vl.DSName(index); dsname != "value" {
		name += "_" + dsname
	}

	switch vl.Dstypes[index] {
	case "counter", "derive":
		name += "_total"
	}

	return metricNameRe.ReplaceAllString(name, "_")
}

// newLabels converts the plugin and type instance of vl to a set of prometheus.Labels.
func newLabels(vl Collectd) prometheus.Labels {
	labels := prometheus.Labels{}
	if vl.PluginInstance != "" {
		labels[vl.Plugin] = vl.PluginInstance
	}
	if vl.TypeInstance != "" {
		if vl.PluginInstance == "" {
			labels[vl.Plugin] = vl.TypeInstance
		} else {
			labels["type"] = vl.TypeInstance
		}
	}
	labels["instance"] = vl.Host

	return labels
}

//newDesc converts one data source of a value list to a Prometheus description.
func newDesc(vl Collectd, index int) *prometheus.Desc {
	help := fmt.Sprintf("Service Assurance Collectd: '%s' Type: '%s' Dstype: '%s' Dsname: '%s'",
		vl.Plugin, vl.Type, vl.Dstypes[index], vl.DSName(index))

	return prometheus.NewDesc(NewName(vl, index), help, []string{}, newLabels(vl))
}

//NewMetric converts one data source of a value list to a Prometheus metric.
func NewMetric(vl Collectd, index int) (prometheus.Metric, error) {
	var value float64
	var valueType prometheus.ValueType

	//switch v := vl.Values[index].(type) {
	switch vl.Dstypes[index] {
	case "gauge":
		value = float64(vl.Values[index])
		valueType = prometheus.GaugeValue
	case "derive":
		value = float64(vl.Values[index])
		valueType = prometheus.CounterValue
	case "counter":
		value = float64(vl.Values[index])
		valueType = prometheus.CounterValue
	default:
		return nil, fmt.Errorf("unknowdsnamen value type: %s", vl.Dstypes[index])
	}

	return prometheus.NewConstMetric(newDesc(vl, index), valueType, value)
}
