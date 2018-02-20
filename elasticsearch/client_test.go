package saelastic

import (

	"log"
  "github.com/redhat-nfvpe/service-assurance-poc/elasticsearch/mapping"
	"testing"
)

const connectivitydata = `[{"labels":{"alertname":"collectd_connectivity_gauge","instance":"d60b3c68f23e","connectivity":"eno2","type":"interface_status","severity":"OKAY","service":"collectd"},"annotations":{"summary":"","ves":{"domain":"stateChange","eventId":2,"eventName":"interface eno2 up","lastEpochMicrosec":1518188764024922,"priority":"high","reportingEntityName":"collectd connectivity plugin","sequence":0,"sourceName":"eno2","startEpochMicrosec":1518188755700851,"version":1.0,"stateChangeFields":{"newState":"inService","oldState":"outOfService","stateChangeFieldsVersion":1.0,"stateInterface":"eno2"}}},"startsAt":"2018-02-09T15:06:04.024859063Z"}]`

const connectivityDirty =  "[{\"labels\":{\"alertname\":\"collectd_connectivity_gauge\",\"instance\":\"d60b3c68f23e\",\"connectivity\":\"eno2\",\"type\":\"interface_status\",\"severity\":\"FAILURE\",\"service\":\"collectd\"},\"annotations\":{\"summary\":\"\",\"ves\":\"{\\\"domain\\\":\\\"stateChange\\\",\\\"eventId\\\":11,\\\"eventName\\\":\\\"interface eno2 down\\\",\\\"lastEpochMicrosec\\\":1518790014024924,\\\"priority\\\":\\\"high\\\",\\\"reportingEntityName\\\":\\\"collectd connectivity plugin\\\",\\\"sequence\\\":0,\\\"sourceName\\\":\\\"eno2\\\",\\\"startEpochMicrosec\\\":1518790009881440,\\\"version\\\":1.0,\\\"stateChangeFields\\\":{\\\"newState\\\":\\\"outOfService\\\",\\\"oldState\\\":\\\"inService\\\",\\\"stateChangeFieldsVersion\\\":1.0,\\\"stateInterface\\\":\\\"eno2\\\"}}\"},\"startsAt\":\"2018-02-16T14:06:54.024856417Z\"}]"

const procEventData=`[{"labels":{"alertname":"collectd_procevent_gauge","instance":"d60b3c68f23e","procevent":"bla.py","type":"process_status","severity":"FAILURE","service":"collectd"},"annotations":{"summary":"","ves":{"domain":"fault","eventId":3,"eventName":"process bla.py (8537) down","lastEpochMicrosec":1518791119579620,"priority":"high","reportingEntityName":"collectd procevent plugin","sequence":0,"sourceName":"bla.py","startEpochMicrosec":1518791111336973,"version":1.0,"faultFields":{"alarmCondition":"process bla.py (8537) state change","alarmInterfaceA":"bla.py","eventSeverity":"CRITICAL","eventSourceType":"process","faultFieldsVersion":1.0,"specificProblem":"process bla.py (8537) down","vfStatus":"Ready to terminate"}}},"startsAt":"2018-02-16T14:25:19.579573212Z"}]`

const procEventDirty="[{\"labels\":{\"alertname\":\"collectd_procevent_gauge\",\"instance\":\"d60b3c68f23e\",\"procevent\":\"bla.py\",\"type\":\"process_status\",\"severity\":\"FAILURE\",\"service\":\"collectd\"},\"annotations\":{\"summary\":\"\",\"ves\":\"{\\\"domain\\\":\\\"fault\\\",\\\"eventId\\\":3,\\\"eventName\\\":\\\"process bla.py (8537) down\\\",\\\"lastEpochMicrosec\\\":1518791119579620,\\\"priority\\\":\\\"high\\\",\\\"reportingEntityName\\\":\\\"collectd procevent plugin\\\",\\\"sequence\\\":0,\\\"sourceName\\\":\\\"bla.py\\\",\\\"startEpochMicrosec\\\":1518791111336973,\\\"version\\\":1.0,\\\"faultFields\\\":{\\\"alarmCondition\\\":\\\"process bla.py (8537) state change\\\",\\\"alarmInterfaceA\\\":\\\"bla.py\\\",\\\"eventSeverity\\\":\\\"CRITICAL\\\",\\\"eventSourceType\\\":\\\"process\\\",\\\"faultFieldsVersion\\\":1.0,\\\"specificProblem\\\":\\\"process bla.py (8537) down\\\",\\\"vfStatus\\\":\\\"Ready to terminate\\\"}}\"},\"startsAt\":\"2018-02-16T14:25:19.579573212Z\"}]"

const elastichost="http://10.19.110.5:9200"
func TestIndexCheckConnectivity(t *testing.T) {
	indexName, indexType, err := GetIndexNameType(connectivitydata)
	if err != nil {
		t.Errorf("Failed to get indexname and type%s", err)
	}
	if indexType != CONNECTIVITYINDEXTYPE {
		t.Errorf("Excepected Index Type %s Got %s", CONNECTIVITYINDEXTYPE, indexType)
	}
	if CONNECTIVITYINDEX != indexName {
		t.Errorf("Excepected Index %s Got %s", CONNECTIVITYINDEX, indexName)
	}

}
func TestSanitize(t *testing.T) {
  result:=Sanitize(procEventDirty)
  log.Println(result)
}


func TestClient(t *testing.T) {
var client *ElasticClient
client=CreateClient(elastichost)
if client.err != nil{
  	t.Errorf("Failed to connect to elastic search%s", client.err)
}
}

func TestIndexCreateAndDelete(t *testing.T) {
  var client *ElasticClient
  client=CreateClient(elastichost)
  if client.err != nil{
    	t.Errorf("Failed to connect to elastic search%s", client.err)
  }else{
     indexName, _, err := GetIndexNameType(connectivitydata)
  	if err != nil {
  		t.Errorf("Failed to get indexname and type%s", err)
      return
  	}

    err=client.DeleteIndex(indexName)

    client.CreateIndex(indexName, saelastic.ConnectivityMapping)
    exists, err := client.client.IndexExists(string(indexName)).Do(client.ctx)
    if (exists==false || err !=nil){
      	t.Errorf("Failed to create index %s", err)
    }
    err=client.DeleteIndex(indexName)
    if ( err !=nil){
      	t.Errorf("Failed to Delete index %s", err)
    }
  }

}

func TestConnectivityDataCreate(t *testing.T) {
  var client *ElasticClient
  client=CreateClient(elastichost)
  if client.err != nil{
      t.Errorf("Failed to connect to elastic search%s", client.err)
  }else{
     indexName, IndexType, err := GetIndexNameType(connectivitydata)
    if err != nil {
      t.Errorf("Failed to get indexname and type%s", err)
      return
    }

    err=client.DeleteIndex(indexName)

    client.CreateIndex(indexName, saelastic.ConnectivityMapping)
    exists, err := client.client.IndexExists(string(indexName)).Do(client.ctx)
    if (exists==false || err !=nil){
        t.Errorf("Failed to create index %s", err)
    }

    id,err:=client.Create(indexName,IndexType,connectivitydata)
    if ( err !=nil){
        t.Errorf("Failed to create data %s\n", err.Error())
    }else{
      log.Printf("document id  %#v\n",id)
    }
    result,err:=client.Get(indexName,IndexType,id)
    if ( err !=nil){
        t.Errorf("Failed to get data %s", err)
    }else{
      log.Printf("Data %#v",result)
    }
    deleteErr:=client.Delete(indexName,IndexType,id)
    if ( deleteErr !=nil){
        t.Errorf("Failed to delete data %s", deleteErr)
    }

    err=client.DeleteIndex(indexName)
    if ( err !=nil){
        t.Errorf("Failed to Delete index %s", err)
    }
  }

}
