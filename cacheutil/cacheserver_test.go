package cacheutil
import (
  "testing"
  "github.com/aneeshkp/service-assurance-goclient/incoming"
  "time"
  //"sync"

)

func TestCacheServer(t *testing.T){
  var pluginCount int =10
  collectd:=incoming.NewInComing(incoming.COLLECTD)
  server:=NewCacheServer(incoming.COLLECTD)
  server.GenrateSampleData("hostname", pluginCount, collectd)
  var incomingDataCache *IncomingDataCache
  incomingDataCache=server.GetCache()
  /*for _, shard := range incomingDataCache.GetShard("hostname"){
      fmt.Println("%#v",plugin )
  }*/
  time.Sleep(time.Second * 1)
  if size :=incomingDataCache.Size();size!=1{
    t.Errorf("wrong count of host , expected 1 and got %d", size)
  }
  if size := incomingDataCache.GetShard("hostname").Size(); size !=pluginCount {
    t.Errorf("wrong count of plugin per host , expected %d and got %d", pluginCount, size)
  }
  server.close()


}
