package saelastic

//import "github.com/elastic/go-elasticsearch/client"
//import "github.com/redhat-nfvpe/service-assurance-poc/elasticsearch/mapping"
import "github.com/olivere/elastic"
import "github.com/satori/go.uuid"

import (
	"context"
	"log"
)

//IndexName   ..
type IndexName string

//IndexType ....
type IndexType string

//COLLECTD
const (
	CONNECTIVITYINDEX IndexName = "collectd_connectivity"
	PROCEVENTINDEX    IndexName = "collectd_procevent"
	SYSEVENTINDEX     IndexName = "collectd_syslogs"
	GENERICINDEX      IndexName = "collectd_generic"
)

//Index Type
const (
	CONNECTIVITYINDEXTYPE IndexType = "event"
	PROCEVENTINDEXTYPE    IndexType = "event"
	EVENTINDEXTYPE        IndexType = "event"
	GENERICINDEXTYPE      IndexType = "event"
)

//ElasticClient  ....
type ElasticClient struct {
	client *elastic.Client
	ctx    context.Context
	err    error
}

//InitAllMappings ....
func (ec *ElasticClient) InitAllMappings() {
	ec.DeleteIndex(string(CONNECTIVITYINDEX))
	ec.DeleteIndex(string(PROCEVENTINDEX))
	ec.DeleteIndex(string(SYSEVENTINDEX))
	ec.DeleteIndex(string(GENERICINDEX))
	//do not create now and leave it for defaults
	/*ec.CreateIndex(string(CONNECTIVITYINDEX), saelastic.ConnectivityMapping)
	ec.CreateIndex(string(PROCEVENTINDEX), saelastic.ConnectivityMapping)
	ec.CreateIndex(string(SYSEVENTINDEX), saelastic.ConnectivityMapping)
	*/
}

//CreateClient   ....
func CreateClient(elastichost string) *ElasticClient {
	//c, _ = client.New(client.WithHosts([]string{"https://elasticsearch:9200"}))
	var elasticClient *ElasticClient
	//var eClient *elastic.Client
	eclient, err := elastic.NewClient(elastic.SetURL(elastichost))
	if err != nil {
		log.Fatal(err)
		elasticClient.err = err
		return elasticClient
	}
	elasticClient = &ElasticClient{client: eclient, ctx: context.Background()}
	elasticClient.InitAllMappings()
	return elasticClient
}

//CreateIndex  ...
func (ec *ElasticClient) CreateIndex(index string, mapping string) {

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

//Create ...  it can be BodyJson or BodyString.. BodyJson needs struct defined
func (ec *ElasticClient) Create(indexname string, indextype IndexType, jsondata string) (string, error) {
	ctx := ec.ctx
	id := genUUIDv4()
	body := Sanitize(jsondata)
	log.Printf("Printing body %s\n", body)
	result, err := ec.client.Index().
		Index(string(indexname)).
		Type(string(indextype)).
		Id(id).
		BodyString(body).
		Do(ctx)
	if err != nil {
		// Handle error
		return id, err
	}
	log.Printf("Indexed  %s to index %s, type %s\n", result.Id, result.Index, result.Type)
	// Flush to make sure the documents got written.
	// Flush asks Elasticsearch to free memory from the index and
	// flush data to disk.
	_, err = ec.client.Flush().Index(string(indexname)).Do(ctx)
	return id, err

}

//Update ....
func (ec *ElasticClient) Update() {

}

//DeleteIndex ...
func (ec *ElasticClient) DeleteIndex(index string) error {
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
func (ec *ElasticClient) Delete(indexname string, indextype IndexType, id string) error {
	// Get tweet with specified ID

	_, err := ec.client.Delete().
		Index(string(indexname)).
		Type(string(indextype)).
		Id(id).
		Do(ec.ctx)
	return err
}

//Get  ....
func (ec *ElasticClient) Get(indexname string, indextype IndexType, id string) (*elastic.GetResult, error) {
	// Get tweet with specified ID

	result, err := ec.client.Get().
		Index(string(indexname)).
		Type(string(indextype)).
		Id(id).
		Do(ec.ctx)
	if err != nil {
		// Handle error
		return nil, err
	}
	/*if result.Found {
		return result.Fields,nil
	}*/
	if result.Found {
		log.Printf("Got document %s in version %d from index %s, type %s\n", result.Id, result.Version, result.Index, result.Type)
	}
	return result, nil
}

//Search  ..
func (ec *ElasticClient) Search(indexname string) *elastic.SearchResult {
	// Search with a term

	termQuery := elastic.NewTermQuery("user", "olivere")
	searchResult, err := ec.client.Search().
		Index(indexname).   // search in index "twitter"
		Query(termQuery).   // specify the query
		Sort("user", true). // sort by "user" field, ascending
		From(0).Size(10).   // take documents 0-9
		Pretty(true).       // pretty print request and response JSON
		Do(ec.ctx)          // execute
	if err != nil {
		// Handle error
		panic(err)
	}
	log.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
	return searchResult

}
