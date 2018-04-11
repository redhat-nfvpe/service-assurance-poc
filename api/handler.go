package apihandler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/redhat-nfvpe/service-assurance-poc/amqp"
	"github.com/redhat-nfvpe/service-assurance-poc/config"
)

type (

	// Timestamp is a helper for (un)marhalling time
	Timestamp time.Time

	// HookMessage is the message we receive from Alertmanager
	HookMessage struct {
		Version           string            `json:"version"`
		GroupKey          string            `json:"groupKey"`
		Status            string            `json:"status"`
		Receiver          string            `json:"receiver"`
		GroupLabels       map[string]string `json:"groupLabels"`
		CommonLabels      map[string]string `json:"commonLabels"`
		CommonAnnotations map[string]string `json:"commonAnnotations"`
		ExternalURL       string            `json:"externalURL"`
		Alerts            []Alert           `json:"alerts"`
	}

	// Alert is a single alert.
	Alert struct {
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		StartsAt    string            `json:"startsAt,omitempty"`
		EndsAt      string            `json:"EndsAt,omitempty"`
	}
	ApiContext struct {
		Config      *saconfig.EventConfiguration
		AMQP1Sender *amqp10.AMQPSender

		// ... and the rest of our globals.
	}
	Handler struct {
		*ApiContext
		H func(c *ApiContext, w http.ResponseWriter, r *http.Request) (int, error)
	}
	body_struct struct {
		body string
	}
)

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok\n")
}

func (ah Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Updated to pass ah.appContext as a parameter to our handler type.
	log.Printf("API is invoked")
	status, err := ah.H(ah.ApiContext, w, r)
	if err != nil {
		log.Printf("HTTP %d: %q", status, err)
		switch status {
		case http.StatusNotFound:
			http.NotFound(w, r)
			// And if we wanted a friendlier error page:
			// err := ah.renderTemplate(w, "http_404.tmpl", nil)
		case http.StatusInternalServerError:
			http.Error(w, http.StatusText(status), status)
		default:
			http.Error(w, http.StatusText(status), status)
		}
	}
}

func AlertHandler(a *ApiContext, w http.ResponseWriter, r *http.Request) (int, error) {
	log.Printf("AlertHandler is invoked")
	log.Printf("BODY %#v", r.Body)
	var body HookMessage
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&body); err != nil {
		log.Printf("error decoding message: %v", err)
		http.Error(w, "invalid request body", 400)
		return http.StatusInternalServerError, err
	}

	log.Printf("body.body %#v\n", body)
	out, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	a.AMQP1Sender.Send(string(out))

	// We can shortcut this: since renderTemplate returns `error`,
	// our ServeHTTP method will return a HTTP 500 instead and won't
	// attempt to write a broken template out with a HTTP 200 status.
	// (see the postscript for how renderTemplate is implemented)
	// If it doesn't return an error, things will go as planned.
	return http.StatusOK, nil
}
