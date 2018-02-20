package main

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/redhat-nfvpe/service-assurance-poc/amqp"
	"github.com/redhat-nfvpe/service-assurance-poc/config"
	"github.com/redhat-nfvpe/service-assurance-poc/elasticsearch"

	"flag"
	"fmt"
	"log"
	"os"
)

/*************** main routine ***********************/
// Usage and command-line flags
func usage() {
	doc := heredoc.Doc(`
  For running with config file use
	********************* config *********************
	$go run events/main.go -config sa.events.config.json
	**************************************************
	For running with AMQP and Prometheus use following option
	********************* Production *********************
	$go run events/main.go -amqp1EventURL=10.19.110.5:5672/collectd/notify -eshost=http://10.19.110.5:9200
	**************************************************************`)
	fmt.Fprintln(os.Stderr, `Required commandline argument missing`)
	fmt.Fprintln(os.Stdout, doc)
	flag.PrintDefaults()
}

func main() {
	// set flags for parsing options
	flag.Usage = usage
	fConfigLocation := flag.String("config", "", "Path to configuration file(optional).if provided ignores all command line options")

	fAMQP1EventURL := flag.String("amqp1EventURL", "", "AMQP1.0 events listener example 127.0.0.1:5672/collectd/notify")
	fElasticHostURL := flag.String("eshost", "", "elasticsearch host http://localhost:9200")

	flag.Parse()
	var serverConfig saconfig.EventConfiguration
	if len(*fConfigLocation) > 0 { //load configuration
		serverConfig = saconfig.LoadEventConfig(*fConfigLocation)
	} else {
		serverConfig = saconfig.EventConfiguration{
			AMQP1EventURL: *fAMQP1EventURL,
			ElasticHostURL: *fElasticHostURL,
		}

	}

	if len(serverConfig.AMQP1EventURL) == 0 {
		log.Println("AMQP1 Event URL is required")
		usage()
		os.Exit(1)
	}
	if len(serverConfig.ElasticHostURL) == 0 {
		log.Println("Elastic Host URL is required")
		usage()
		os.Exit(1)
	}

	eventsNotifier := make(chan string) // Channel for messages from goroutines to main()
	var amqpEventServer *amqplistener.AMQPServer
	///Metric Listener
	amqpEventsurl := fmt.Sprintf("amqp://%s", serverConfig.AMQP1EventURL)
	amqpEventServer = amqplistener.NewAMQPServer(amqpEventsurl, true, -1, eventsNotifier)
	var elasticClient *saelastic.SAElasticClient
	elasticClient = saelastic.CreateClient(serverConfig.ElasticHostURL)

	for {
		select {
		case event := <-amqpEventServer.GetNotifier():
			//log.Printf("Event occured : %#v\n", event)
			indexName, indexType, err := saelastic.GetIndexNameType(event)
			if err != nil {
				log.Printf("Failed to read event %s type in main %s\n", event,err)
			} else {
				id, err := elasticClient.Create(indexName, indexType, event)
				if err != nil {
					log.Printf("Error creating event %s in elastic search %s\n", event, err)
				} else {
					log.Printf("Document created in elasticsearch for mapping: %s ,type: %s, id :%s\n", string(indexName), string(indexType), id)
				}

			}
			continue
		default:
			//no activity
		}
	}

	//TO DO: to close cache server on keyboard interrupt

}
