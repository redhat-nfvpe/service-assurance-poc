package saelastic

//import "github.com/elastic/go-elasticsearch/client"
import "github.com/redhat-nfvpe/service-assurance-poc/elasticsearch/mapping"
import "github.com/olivere/elastic"
import "github.com/satori/go.uuid"


import (
	"log"
  "context"
)
//IndexName   ..
type IndexName string
//IndexType ....
type IndexType string

//COLLECTD
const (
	CONNECTIVITYINDEX IndexName = "connectivity"
  PROCEVENTINDEX IndexName = "procevent"
  SYSLOGINDEX IndexName = "syslogs"
  GENERICINDEX IndexName = "generic"
)
//Index Type
const (
	CONNECTIVITYINDEXTYPE IndexType = "event"
  PROCEVENTINDEXTYPE IndexType = "event"
  GENERICINDEXTYPE IndexType = "event"
)
//Client  ....
type SAElasticClient struct {
	client *elastic.Client
  ctx context.Context
  err error

}
//InitAllMappings ....
func (ec * SAElasticClient) InitAllMappings(){
  ec.CreateIndex(CONNECTIVITYINDEX,saelastic.ConnectivityMapping)
}

//CreateClient   ....
func CreateClient(elastichost string) *SAElasticClient {
	//c, _ = client.New(client.WithHosts([]string{"https://elasticsearch:9200"}))
	var elasticClient *SAElasticClient
	//var eClient *elastic.Client
	eclient, err := elastic.NewClient(elastic.SetURL(elastichost))
	if err != nil {
		log.Fatal(err)
    elasticClient.err=err
    return elasticClient
	}
	elasticClient = &SAElasticClient{client: eclient,ctx:context.Background()}
  elasticClient.InitAllMappings()
	return elasticClient
}

//CreateIndex  ...
func (ec *SAElasticClient) CreateIndex(index IndexName,mapping string) {

	exists, err := ec.client.IndexExists(string(index)).Do(ec.ctx)
	if err != nil {
		// Handle error nothing to do index exists
	}
	if !exists {
		// Index does not exist yet.
		// Create a new index.
		createIndex, err := ec.client.CreateIndex(string(index)).BodyString(mapping).Do(ec.ctx)
		if err != nil {
			// Handle error
			panic(err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
      log.Println("Index Not acknowledged")
		}
	}

}
//genUUIDv4   ...
func genUUIDv4() string {
	id, _ := uuid.NewV4()
	log.Printf("github.com/satori/go.uuid:   %s\n", id)
  return id.String()
}

//Create...  it can be BodyJson or BodyString.. BodyJson needs struct defined
func (ec *SAElasticClient) Create(indexname IndexName,indextype IndexType,jsondata string) (string, error) {
  ctx :=ec.ctx
  id:=genUUIDv4()
  body:=Sanitize(jsondata)
  log.Printf("Printing body %s\n",body)
  put2, err := ec.client.Index().
  		Index(string(indexname)).
  		Type(string(indextype)).
  		Id(id).
  		BodyString(body).
  		Do(ctx)
  	if err != nil {
  		// Handle error
  		return id,err
  	}
  	log.Printf("Indexed  %s to index %s, type %s\n", put2.Id, put2.Index, put2.Type)
    // Flush to make sure the documents got written.
    // Flush asks Elasticsearch to free memory from the index and
    // flush data to disk.
	  _, err = ec.client.Flush().Index(string(indexname)).Do(ctx)
 		return id,err

}

//Update ....
func (ec *SAElasticClient) Update() {

}

//Delete
func (ec *SAElasticClient) DeleteIndex(index IndexName) error {
  // Delete an index.
  	deleteIndex, err := ec.client.DeleteIndex(string(index)).Do(ec.ctx)
  	if err != nil {
  		// Handle error
  		//panic(err)
      return err
  	}
  	if !deleteIndex.Acknowledged {
  		// Not acknowledged
  	}
    return nil
}


//Delete  ....
func (ec *SAElasticClient) Delete(indexname IndexName,indextype IndexType,id string) error {
  // Get tweet with specified ID

	_,err := ec.client.Delete().
		Index(string(indexname)).
		Type(string(indextype)).
		Id(id).
		Do(ec.ctx)
  return err
}

//Get  ....
func (ec *SAElasticClient) Get(indexname IndexName,indextype IndexType,id string)(*elastic.GetResult,error) {
  // Get tweet with specified ID

	result,err := ec.client.Get().
		Index(string(indexname)).
		Type(string(indextype)).
		Id(id).
		Do(ec.ctx)
	if err != nil {
		// Handle error
		return nil,err
	}
	/*if result.Found {
		return result.Fields,nil
	}*/
  if result.Found {
		log.Printf("Got document %s in version %d from index %s, type %s\n", result.Id, result.Version, result.Index, result.Type)
	}
  return result,nil
}

//Search  ..
func (ec *SAElasticClient) Search(indexname string) *elastic.SearchResult {
  // Search with a term

	termQuery := elastic.NewTermQuery("user", "olivere")
	searchResult, err := ec.client.Search().
		Index(indexname).   // search in index "twitter"
		Query(termQuery).   // specify the query
		Sort("user", true). // sort by "user" field, ascending
		From(0).Size(10).   // take documents 0-9
		Pretty(true).       // pretty print request and response JSON
		Do(ec.ctx)             // execute
	if err != nil {
		// Handle error
		panic(err)
	}
  log.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
	return searchResult

}
