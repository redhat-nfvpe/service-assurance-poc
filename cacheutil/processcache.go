package cacheutil

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redhat-nfvpe/service-assurance-poc/incoming"
	"github.com/redhat-nfvpe/service-assurance-poc/tsdb"
	"log"
)

//FlushPrometheusMetric   generate Prometheus metrics
func (shard *ShardedIncomingDataCache) FlushPrometheusMetric(ch chan<- prometheus.Metric) {
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
			} else {
				//clean up if data is not access for max TTL specified
				if shard.Expired() {
					delete(shard.plugin, collectd.GetItemKey())
					//log.Printf("Cleaned up plugin for %s", collectd.GetKey())
				}
			}
		}
	}
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
