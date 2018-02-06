/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"qpid.apache.org/amqp"
	"qpid.apache.org/electron"
	"strings"
	"sync"
	"time"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s (options) ampq://... \n", os.Args[0])
	fmt.Fprintf(os.Stderr, "options:\n")
	flag.PrintDefaults()
}

var hostnameTemplate = "hostname%03d"
var metricsTemplate = "metrics%03d"

type metric struct {
	hostname *string
	name     string
	interval int
}

type host struct {
	name    string
	metrics []metric
}

func (m *metric) GetMetricMessage(i int) (msg string) {
	msgTemplate := `
[{"values": [%f], "dstypes": ["derive"], "dsnames": ["samples"],
"time": %f, "interval": 10, "host": "%s", "plugin": "testPlugin",
"plugin_instance": "testInstance","type": "%v","type_instance": ""}]
`
	msg = fmt.Sprintf(msgTemplate,
		rand.Float64(),                           // val
		float64((time.Now().UnixNano()))/1000000, // time
		*m.hostname,                              // host
		m.name)                                   // type

	return
}

func generateHosts(hostsNum int, metricNum int, intervalSec int) []host {

	hosts := make([]host, hostsNum)
	for i := 0; i < hostsNum; i++ {
		hosts[i].name = fmt.Sprintf(hostnameTemplate, i)
		hosts[i].metrics = make([]metric, metricNum)
		for j := 0; j < metricNum; j++ {
			hosts[i].metrics[j].name =
				fmt.Sprintf(metricsTemplate, j)
			hosts[i].metrics[j].interval = intervalSec
			hosts[i].metrics[j].hostname = &hosts[i].name
		}
	}
	return hosts
}

func printHostsInfo(hosts *[]host) {
	for _, v := range *hosts {
		for _, w := range v.metrics {
			fmt.Printf("%v.%v\n", v.name, w.name)
		}
	}
}

func main() {
	// ./sa-bench -hosts 3 -metrics 2 -send 7 amqp://localhost:5672/foo

	hostsNum := flag.Int("hosts", 1, "Number of hosts to simulate")
	metricsNum := flag.Int("metrics", 1, "Metrics per hosts")
	intervalSec := flag.Int("interval", 1, "interval (sec)")
	metricMaxSend := flag.Int("send", 1, "How many metrics sent")

	flag.Usage = usage
	flag.Parse()

	urls := flag.Args()
	if len(urls) == 0 {
		fmt.Fprintln(os.Stderr, "ampq URL is missing")
		usage()
		os.Exit(1)
	} else if len(urls) > 1 {
		fmt.Fprintln(os.Stderr, "Only one ampq URL is supported")
		usage()
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())
	hosts := generateHosts(*hostsNum, *metricsNum, *intervalSec)
	//printHostsInfo(&hosts)

	container := electron.NewContainer(fmt.Sprintf("sabe%d", os.Getpid()))
	url, err := amqp.ParseURL(urls[0])
	if err != nil {
		log.Fatal(err)
		return
	}

	con, err := container.Dial("tcp", url.Host)
	if err != nil {
		log.Fatal(err)
		return
	}

	ackChan := make(chan electron.Outcome)

	var wait sync.WaitGroup
	var waitb sync.WaitGroup
	for _, v := range hosts {
		for _, w := range v.metrics {
			// uncomment if need to rondom wait
			/*
				time.Sleep(time.Millisecond *
					time.Duration(rand.Int()%1000))
			*/
			wait.Add(1)
			go func(m metric) {
				defer wait.Done()

				addr := strings.TrimPrefix(url.Path, "/")
				s, err := con.Sender(electron.Target(addr))
				if err != nil {
					log.Fatal(err)
				}

				for i := 0; ; i++ {
					if i >= *metricMaxSend &&
						*metricMaxSend != -1 {
						break
					}

					msg := amqp.NewMessage()
					body := m.GetMetricMessage(i)
					msg.Marshal(body)
					s.SendAsync(msg, ackChan, body)
					fmt.Printf("sent: H:%v M:%v (%d)\n", *m.hostname, m.name, i)
					time.Sleep(time.Duration(m.interval) * time.Second)
				}
			}(w)
		}
	}

	cancel := make(chan struct{})
	// routine for waiting ack....
	waitb.Add(1)
	go func() {
		for {
			select {
			case out := <-ackChan:
				if out.Error != nil {
					log.Fatalf("acknowledgement %v error: %v",
						out.Value, out.Error)
				} else if out.Status != electron.Accepted {
					//log.Fatalf("acknowledgement unexpected status: %v", out.Status)
					log.Printf("acknowledgement unexpected status: %v", out.Status)
				}
				/*
					} else {
						fmt.Printf("acknowledgement %v (%v)\n",
							out.Value, out.Status)
				*/
			case <-cancel:
				waitb.Done()
				return
			}
		}
	}()

	wait.Wait()
	close(cancel)
	waitb.Wait()
	con.Close(nil)
}
