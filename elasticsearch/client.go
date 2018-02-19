package elastic

//import "github.com/elastic/go-elasticsearch/client"
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
func (ec *SAElasticClient) Create(indexname string,indextype string,jsondata string) {
  ctx :=ec.ctx
  put2, err := ec.client.Index().
  		Index(indexname).
  		Type(indextype).
  		Id(genUUIDv4()).
  		BodyString(jsondata).
  		Do(ctx)
  	if err != nil {
  		// Handle error
  		panic(err)
  	}
  	log.Printf("Indexed tweet %s to index %s, type %s\n", put2.Id, put2.Index, put2.Type)
    // Flush to make sure the documents got written.
	  _, err = ec.client.Flush().Index(indexname).Do(ctx)
	  if err != nil {
   		panic(err)
	 }
}

//Update ....
func (ec *SAElasticClient) Update() {

}

//Delete
func (ec *SAElasticClient) Delete(index string) {
  // Delete an index.
  	deleteIndex, err := ec.client.DeleteIndex(index).Do(ec.ctx)
  	if err != nil {
  		// Handle error
  		panic(err)
  	}
  	if !deleteIndex.Acknowledged {
  		// Not acknowledged
  	}
}

//Search  ..
func (ec *SAElasticClient) Search(indexname string) *elastic.SearchResult {
  // Search with a term query
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
