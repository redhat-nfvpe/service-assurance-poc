package webserver

import (
	"context"

	"log"
	"net/http"
	"sync"
)

// WebServer defines the handler endpoints and launches the web server listener
func WebServer(host string, handler http.Handler, wg *sync.WaitGroup, shutdown chan struct{}) {

	// Make sure the exit is noted
	//ctx  context.Context
	ctx, _ := context.WithCancel(context.Background())

	defer wg.Done()

	// Go forth and serve
	// Define the server struct
	srv := &http.Server{
		Addr:    host,
		Handler: handler,
	}
	log.Println("2. Starting up Web Server.")

	// Now go serve
	log.Println("Launching web server.")
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		err := srv.ListenAndServe()
		if err != nil &&
			err.Error() != "http: Server closed" {

			// There...was an error
			log.Println("Error launching web server: " + err.Error())
		}
	}(wg)

	// Listen for a shutdown
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case <-shutdown:
				// Shut down the process
				log.Println("Shutting down web server process.")
				if err := srv.Shutdown(ctx); err != nil {

					// There was a problem shutting down gracefully
					log.Println("Error shutting down web server: " + err.Error())
					return
				}

				// Done
				log.Println("Web server has shutdown.")
				return

			default:
				continue
			}
		}
	}(wg)

	// Done
	return
}
