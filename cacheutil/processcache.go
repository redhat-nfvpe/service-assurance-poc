package cacheutil

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redhat-nfvpe/service-assurance-poc/incoming"
	"github.com/redhat-nfvpe/service-assurance-poc/tsdb"
)

//AddHeartBeat ...
func AddHeartBeat(instance string, ch chan<- prometheus.Metric) {
	m, err := tsdb.NewHeartBeatMetric(instance)
	if err != nil {
		log.Printf("newHeartBeat: %v for %s", err, instance)
	}
	ch <- m
}

//FlushPrometheusMetric   generate Prometheus metrics
func (shard *ShardedIncomingDataCache) FlushPrometheusMetric(ch chan<- prometheus.Metric) bool {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	minMetricCreated := false //..minimum of one metrics created
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
					minMetricCreated = true
				}
			} else {
				//clean up if data is not access for max TTL specified
				if shard.Expired() {
					delete(shard.plugin, collectd.GetItemKey())
					//log.Printf("Cleaned up plugin for %s", collectd.GetKey())
				}
			}
		}
	}
	return minMetricCreated
}

//FlushAllMetrics   Generic Flushing metrics not used.. used only for testing
func (shard *ShardedIncomingDataCache) FlushAllMetrics() {
	shard.lock.Lock()
	defer shard.lock.Unlock()
	for _, IncomingDataInterface := range shard.plugin {
		if collectd, ok := IncomingDataInterface.(*incoming.Collectd); ok {
			if collectd.ISNew() {
				collectd.SetNew(false)
				log.Printf("New Metrics %#v\n", collectd)
			} else {
				//clean up if data is not access for max TTL specified
				if shard.Expired() {
					delete(shard.plugin, collectd.GetItemKey())
					log.Printf("Cleaned up plugin for %s", collectd.GetItemKey())
				}
			}
		}
	}
}
