package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"os/signal"
	"time"

	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/MakeNowJust/heredoc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redhat-nfvpe/service-assurance-poc/alerts"
	"github.com/redhat-nfvpe/service-assurance-poc/amqp"
	"github.com/redhat-nfvpe/service-assurance-poc/api"
	"github.com/redhat-nfvpe/service-assurance-poc/cacheutil"
	"github.com/redhat-nfvpe/service-assurance-poc/config"
	"github.com/redhat-nfvpe/service-assurance-poc/elasticsearch"
	"github.com/redhat-nfvpe/service-assurance-poc/webserver"
)

var (
	shutdown     = make(chan struct{})
	done         = make(chan bool)
	debuge       = func(format string, data ...interface{}) {} // Default no debugging output
	m1           runtime.MemStats
	m2           runtime.MemStats
	serverConfig saconfig.EventConfiguration
	wg           sync.WaitGroup
)

func closeAll() {
	close(shutdown)
	log.Println("Sending shutdown signal..")
}

/*************** main routine ***********************/
// eventusage and command-line flags
func eventusage() {
	doc := heredoc.Doc(`
  For running with config file use
	********************* config *********************
	$go run events/main.go -config sa.events.config.json -debug
	**************************************************
	For running with AMQP and Prometheus use following option
	********************* Production *********************
	$go run events/main.go -amqp1EventURL=10.19.110.5:5672/collectd/notify -eshost=http://10.19.110.5:9200
	**************************************************************
	For running with AMQP ,Prometheus,API and AlertManager use following option
	********************* Production *********************
	$go run events/main.go -amqp1EventURL=10.19.110.5:5672/collectd/notify -eshost=http://10.19.110.5:9200 -alertmanager=http://localhost:9090/v1/api/alert -apiurl=localhost:8082 -amqppublishurl=127.0.0.1:5672/collectd/alert
	**************************************************************`)
	fmt.Fprintln(os.Stderr, `Required commandline argument missing`)
	fmt.Fprintln(os.Stdout, doc)
	flag.PrintDefaults()
}

func printMemStats() {
	runtime.ReadMemStats(&m2)
	debuge("Number of goroutines: %d\n", runtime.NumGoroutine())
	debuge("Memory: %.2f bytes\n", float64(m2.Sys-m1.Sys)/float64(runtime.NumGoroutine()))
}
func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	runtime.ReadMemStats(&m1)
	printMemStats()

	// set flags for parsing options
	flag.Usage = eventusage
	fDebug := flag.Bool("debug", false, "Enable debug")
	fConfigLocation := flag.String("config", "", "Path to configuration file(optional).if provided ignores all command line options")
	fAMQP1EventURL := flag.String("amqp1EventURL", "", "AMQP1.0 events listener example 127.0.0.1:5672/collectd/notify")
	fElasticHostURL := flag.String("eshost", "", "ElasticSearch host http://localhost:9200")
	fAlertManagerURL := flag.String("alertmanager", "", "(Optional)AlertManager endpoint http://localhost:9090/v1/api/alert")
	fAPIEndpointURL := flag.String("apiurl", "", "(Optional)API endpoint localhost:8082")
	fAMQP1PublishURL := flag.String("amqppublishurl", "", "(Optional) AMQP1.0 event publish address 127.0.0.1:5672/collectd/alert")
	fResetIndex := flag.Bool("resetIndex", false, "Optional Clean all index before on start (default false)")

	flag.Parse()
	if len(*fConfigLocation) > 0 { //load configuration
		serverConfig = saconfig.LoadEventConfig(*fConfigLocation)
		if *fDebug {
			serverConfig.Debug = true
		}
	} else {
		serverConfig = saconfig.EventConfiguration{
			AMQP1EventURL:   *fAMQP1EventURL,
			ElasticHostURL:  *fElasticHostURL,
			AlertManagerURL: *fAlertManagerURL,
			API: saconfig.EventAPIConfig{
				APIEndpointURL:  *fAPIEndpointURL,
				AMQP1PublishURL: *fAMQP1PublishURL,
			},
			ResetIndex: *fResetIndex,
			Debug:      *fDebug,
		}

	}
	if serverConfig.Debug {
		debuge = func(format string, data ...interface{}) { log.Printf(format, data...) }
	}

	if len(serverConfig.AMQP1EventURL) == 0 {
		log.Println("AMQP1 Event URL is required")
		eventusage()
		os.Exit(1)
	}
	if len(serverConfig.ElasticHostURL) == 0 {
		log.Println("Elastic Host URL is required")
		eventusage()
		os.Exit(1)
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	printMemStats()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for sig := range signalCh {
			log.Printf("caught sig: %+v", sig)
			printMemStats()
			time.Sleep(2 * time.Second)
			closeAll()
			printMemStats()
			//os.Exit(0)
			done <- true
			break
		}
	}()

	if len(serverConfig.AlertManagerURL) > 0 {
		log.Printf("AlertManager configured at %s\n", serverConfig.AlertManagerURL)
		serverConfig.AlertManagerEnabled = true
	} else {
		log.Println("AlertManager disabled")
	}
	if len(serverConfig.API.APIEndpointURL) > 0 {
		log.Printf("API availble at %s\n", serverConfig.API.APIEndpointURL)
		serverConfig.APIEnabled = true
	} else {
		log.Println("API disabled")
	}
	if len(serverConfig.API.AMQP1PublishURL) > 0 {
		log.Printf("AMQP1.0 Publish address at %s\n", serverConfig.API.AMQP1PublishURL)
		serverConfig.PublishEventEnabled = true
	} else {
		log.Println("AMQP1.0 Publish address disabled")
	}

	/* Print Configuration detials */
	//mertic handler for event mertics to check health status
	applicationHealth := cacheutil.NewApplicationHealthCache()
	metricHandler := apihandler.NewAppStateEventMetricHandler(applicationHealth)
	debuge("Debug:Config %#v\n", serverConfig)

	var amqpEventServer *amqp10.AMQPServer
	///Metric Listener
	amqpEventsurl := fmt.Sprintf("amqp://%s", serverConfig.AMQP1EventURL)
	log.Printf("Connecting to AMQP1 : %s\n", amqpEventsurl)
	amqpEventServer = amqp10.NewAMQPServer(amqpEventsurl, serverConfig.Debug, -1)

	log.Printf("Listening.....\n")
	var elasticClient *saelastic.ElasticClient
	log.Printf("Connecting to ElasticSearch : %s\n", serverConfig.ElasticHostURL)
	elasticClient = saelastic.CreateClient(serverConfig.ElasticHostURL, serverConfig.ResetIndex, serverConfig.Debug)
	applicationHealth.ElasticSearchState = 1

	/**** HTTP Listener for alerts from alert manager *******************************
	*
	*
	********************************************************************************/
	//configure http alert route to amqp1.0
	if serverConfig.APIEnabled {
		prometheus.MustRegister(metricHandler)
		// Including these stats kills performance when Prometheus polls with multiple targets
		prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
		prometheus.Unregister(prometheus.NewGoCollector())
		prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
		prometheus.Unregister(prometheus.NewGoCollector())

		context := apihandler.NewAPIContext(serverConfig)
		handler := http.NewServeMux()
		handler.Handle("/alert", apihandler.Handler{context, apihandler.AlertHandler}) //creates writer everytime api is called.
		handler.Handle("/metrics", prometheus.Handler())
		handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html>
																	<head><title>Smart Gateway Event API</title></head>
																	<body>
																	<h1>APi </h1>
																	/alerts Post alerts in Json Format on to amqp bus</br>
																	/metrics get metric data
																	</body>
																	</html>`))
		})
		// Register pprof handlers
		handler.HandleFunc("/debug/pprof/", pprof.Index)
		handler.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		handler.HandleFunc("/debug/pprof/profile", pprof.Profile)
		handler.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		handler.HandleFunc("/debug/pprof/trace", pprof.Trace)

		APIEndpointURL := fmt.Sprintf("%s", serverConfig.API.APIEndpointURL)
		wg.Add(1)
		webserver.WebServer(APIEndpointURL, handler, &wg, shutdown)
		log.Printf("Event API server at : %s\n", APIEndpointURL)
		time.Sleep(2 * time.Second)
		log.Println("HTTP server is ready....")
		printMemStats()
		log.Println("Ready")

	}
	// start processing  events from QDR
LOOP:
	for {
		select {
		case <-shutdown:
			log.Println("Closing Incoming message listener")
			break LOOP
		case event := <-amqpEventServer.GetNotifier():
			//log.Printf("Event occured : %#v\n", event)
			indexName, indexType, err := saelastic.GetIndexNameType(event)
			if err != nil {
				log.Printf("Failed to read event %s type in main %s\n", event, err)
				applicationHealth.ElasticSearchState = 0
			} else {
				id, err := elasticClient.Create(indexName, indexType, event)
				applicationHealth.ElasticSearchState = 1
				if err != nil {
					log.Printf("Error creating event %s in elastic search %s\n", event, err)
					applicationHealth.ElasticSearchState = 0
				} // else {
				//update AlertManager
				if serverConfig.AlertManagerEnabled {
					wg.Add(1)
					go func() {
						defer wg.Done()
						var alert = &alerts.Alerts{}
						var jsonStr = []byte(event)
						generatorURL := fmt.Sprintf("%s/%s/%s/%s", serverConfig.ElasticHostURL, indexName, indexType, id)
						alert.Parse(jsonStr, generatorURL)
						debuge("Debug:Sending alert..%#v\n", alert)
						debuge("Debug:Generator URL %s\n", generatorURL)
						jsonString, err := json.Marshal(*alert)
						if err != nil {
							panic(err)
						}
						var jsonEvent = []byte("[" + string(jsonString) + "]")
						debuge("Debug:Posting to  %#s\n", serverConfig.AlertManagerURL)
						req, err := http.NewRequest("POST", serverConfig.AlertManagerURL, bytes.NewBuffer(jsonEvent))
						req.Header.Set("X-Custom-Header", "smartgateway")
						req.Header.Set("Content-Type", "application/json")
						client := &http.Client{}
						resp, err := client.Do(req)
						if err != nil {
							panic(err)
						}
						defer resp.Body.Close()
						body, _ := ioutil.ReadAll(resp.Body)
						debuge("Debug:response Status:%s\n", resp.Status)
						debuge("Debug:response Headers:%s\n", resp.Header)
						debuge("Debug:response Body:%s\n", string(body))

					}()
					//}
				}
			}
			continue // priority channel
		case status := <-amqpEventServer.GetStatus():
			applicationHealth.QpidRouterState = status
		default:
			//no activity
		}
	}
	debuge("awaiting signal")
	<-done
	debuge("Done signal recieved")
	printMemStats()
	wg.Wait()
	printMemStats()
	log.Println("Goodbye")
	printMemStats()

}
