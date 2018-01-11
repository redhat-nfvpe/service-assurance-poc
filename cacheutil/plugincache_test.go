package cacheutil
import (
  "fmt"
  "testing"

  "math/rand"
)


func populateCacheWithHosts(count int ,hostname string) *Cache {
  //hostDict:=Cache{}
  var hostDict = NewCache()
  for i:=0;i<count;i++ {
    hostDict.Put(fmt.Sprintf("%s_%d", hostname,i))
  }
  return &hostDict
}

func getLabels(hostname string) Label{
  labels :=Label{}
  labels.Put("instance",hostname)
  //labels.Put("id",  strconv.Itoa(id))
  labels.Put("foo","bar")
  return labels
}
//get 100's of  metric for each host
func setPlugin(hostname string, pluginCache *ShardedPluginCache) {
  // initlaizepluginsDic:=
  //some common name
  pluginNames :=[]string{"interface","network","cpuutilization","memoryused","memoryfree"}
  // 100 plugin
  var plugins[100]string

  // generate 100 difference meteric names
  var j int
  for i:=0;i<20;i++ {
    for _,value:= range pluginNames{
    plugins[j]=fmt.Sprintf("%s_%s_%d", "metric",value,j)
    j++
    }
  }
  //data to types for all
  var data[2]string
  data[0]="rx"
  data[1]="tx"
  //for each host get 100 plugin

  for _, pluginNames:= range plugins{

    plugin:=NewPlugin()
    plugin.metrictype ="guage"
    plugin.name = pluginNames
    labels := getLabels(hostname)
    for key, value :=range labels.items {
      plugin.labels.Put(key,value)
    }
    for _, value := range data {
      plugin.datasource.Put(value,rand.Float64())
    }
    //deference pointer befor sending
    pluginCache.Put(plugin.name,*plugin)

    }
}

func TestPut(t *testing.T){
  cache:=populateCacheWithHosts(100,"redhat.bosoton.nfv")
  if size := cache.Size(); size != 100 {
        t.Errorf("wrong count of hosts, expected 100 and got %d", size)
    }

    cache.Put("redhat.bosoton.nfv_99") //should not add a new one, just change the existing one
    if size := cache.Size(); size != 100 {
        t.Errorf("wrong count, expected 100 and got %d", size)
    }
    cache.Put("redhat.bosoton.nfv_101") //should add it
    if size := cache.Size(); size != 101 {
        t.Errorf("wrong count, expected 1plugins01 and got %d", size)
    }
    //get  plugin

    for hostname,pluginCache:= range *cache{
            setPlugin(hostname,pluginCache )
      if size := pluginCache.Size(); size != 100 {
          t.Errorf("wrong count for plugin, expected 10 and got %d", size)
      }
    }
    for _,pluginCache := range *cache{
        if size :=pluginCache.Size(); size != 100 {
        t.Errorf("wrong count for plugin per hosts, expected 100 and got %d", size)
    }
  }

/*
func TestWriteAndRead(t * testing.T){

}
*/





}
