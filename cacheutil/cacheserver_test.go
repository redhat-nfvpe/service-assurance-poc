package cacheutil

import (
	"github.com/redhat-nfvpe/service-assurance-poc/incoming"

	//"sync"
	"testing"
	"time"
)

func TestCacheServer(t *testing.T) {
	//ch:=  make(chan IncomingBuffer)
	hostname := "host"
	collectd := incoming.NewInComing(incoming.COLLECTD)
	newSample := collectd.GenerateSampleData(hostname, "pg")

	if newSample.GetKey() != hostname {
		t.Errorf("Data Key is not matching , expected %s and got %s", hostname, newSample.GetKey())
	}

}

func TestCacheServer2(t *testing.T) {
	pluginCount := 10
	hostname := "hostname"
	//	var hostCount=1
	//	var freeListToCollectSample = make(chan *IncomingBuffer, 100)

	//  collectd:=incoming.NewInComing(incoming.COLLECTD)
	server := NewCacheServer()
	collectd := incoming.NewInComing(incoming.COLLECTD)
	server.GenrateSampleData(hostname, pluginCount, collectd)

	time.Sleep(time.Second * 2)

	incomingDataCache := server.GetCache()
	if size := incomingDataCache.Size(); size != 1 {
		t.Errorf("wrong count of host , expected 1 and got %d", size)
	}
	if size := incomingDataCache.GetShard(hostname).Size(); size != pluginCount {
		t.Errorf("wrong count of plugin per host , expected %d and got %d", pluginCount, size)
	}

}
