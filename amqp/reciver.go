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

package amqp

import (
	"flag"
	"fmt"
	"log"
	"os"
	"qpid.apache.org/amqp"
	"qpid.apache.org/electron"
	"strings"
	"sync"
)

// Usage and command-line flags
func usage() {
	fmt.Fprintf(os.Stderr, `Usage: %s url [url ...]
Receive messages from all URLs concurrently and print them.
URLs are of the form "amqp://<host>:<port>/<amqp-address>"
`, os.Args[0])
	flag.PrintDefaults()
}


var debugf = func(format string, data ...interface{}) {} // Default no debugging output

//AMQP "amqp://<host>:<port>/<amqp-address>"
func AMQP(urlStr string,debug bool,messages chan amqp.Message) {
	flag.Usage = usage
	flag.Parse()
	if len(urlStr) == 0 {
		log.Println("No URL provided")
		os.Exit(1)
	}



	var wait sync.WaitGroup // Used by main() to wait for all goroutines to end.
	wait.Add(1)     // Wait for one goroutine per URL.

	container := electron.NewContainer(fmt.Sprintf("receive[%v]", os.Getpid()))
	connections := make(chan electron.Connection, 1) // Connections to close on exit

	// Start a goroutine to for each URL to receive messages and send them to the messages channel.
	// main() receives and prints them.

		go func(urlStr string) { // Start the goroutine
			defer wait.Done() // Notify main() when this goroutine is done.
			url, err := amqp.ParseURL(urlStr)
			fatalIf(err)
			c, err := container.Dial("tcp", url.Host) // NOTE: Dial takes just the Host part of the URL
			fatalIf(err)
			connections <- c // Save connection so we can Close() when main() ends
			addr := strings.TrimPrefix(url.Path, "/")
			opts := []electron.LinkOption{electron.Source(addr)}
			/*if *prefetch > 0 {
				opts = append(opts, electron.Capacity(*prefetch), electron.Prefetch(true))
			}*/
			r, err := c.Receiver(opts...)
      fmt.Printf("receive error %v: %v", urlStr, err)
			fatalIf(err)
			// Loop receiving messages and sending them to the main() goroutine
			for {
				if rm, err := r.Receive(); err == nil {
					rm.Accept()
					messages <- rm.Message
				} else if err == electron.Closed {
					return
				} else {
          fmt.Printf("receive error %v: %v", urlStr, err)
					log.Fatalf("receive error %v: %v", urlStr, err)
				}
			}
		}(urlStr)


	// All goroutines are started, we are receiving messages.
	fmt.Printf("Listening on %s connections\n", urlStr)

	// print each message until the count is exceeded.
	//for i := uint64(0); i < *count; i++ {
  /*for{
		m := <-messages
		debugf("%v\n", m.Body())
	}*/
	//fmt.Printf("Received %d messages\n", *count)

	// Close all connections, this will interrupt goroutines blocked in Receiver.Receive()
	// with electron.Closed.
	/*for i := 0; i < len(urls); i++ {
		c := <-connections
		debugf("close %s", c)
		c.Close(nil)
	}*/
	wait.Wait() // Wait for all goroutines to finish.
}

func fatalIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
