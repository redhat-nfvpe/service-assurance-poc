package apihandler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/redhat-nfvpe/service-assurance-poc/amqp"
	"github.com/redhat-nfvpe/service-assurance-poc/config"
)

var debugh = func(format string, data ...interface{}) {} // Default no debugging output

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

	//Alert is a single alert.
	Alert struct {
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		StartsAt    string            `json:"startsAt,omitempty"`
		EndsAt      string            `json:"EndsAt,omitempty"`
	}
	//APIContext ...
	APIContext struct {
		Config      *saconfig.EventConfiguration
		AMQP1Sender *amqp10.AMQPSender
	}
	//Handler ...
	Handler struct {
		*APIContext
		H func(c *APIContext, w http.ResponseWriter, r *http.Request) (int, error)
	}
)

//NewAPIContext ...
func NewAPIContext(serverConfig saconfig.EventConfiguration) *APIContext {
	amqpPublishurl := fmt.Sprintf("amqp://%s", serverConfig.API.AMQP1PublishURL)
	amqpSender := amqp10.NewAMQPSender(amqpPublishurl, false)
	context := &APIContext{Config: &serverConfig, AMQP1Sender: amqpSender}
	if serverConfig.Debug {
		debugh = func(format string, data ...interface{}) { log.Printf(format, data...) }
	}
	return context
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "ok\n")
}

//ServeHTTP...
func (ah Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Updated to pass ah.appContext as a parameter to our handler type.
	status, err := ah.H(ah.APIContext, w, r)
	if err != nil {
		debugh("Debug:HTTP %d: %q", status, err)
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

//AlertHandler  ...
func AlertHandler(a *APIContext, w http.ResponseWriter, r *http.Request) (int, error) {
	var body HookMessage
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&body); err != nil {
		http.Error(w, "invalid request body", 400)
		return http.StatusInternalServerError, err
	}

	debugh("API AlertHandler Body%#v\n", body)
	out, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	debugh("Debug:Sending alerts to to AMQP")
	debugh("Debug:Alert on AMQP%#v\n", string(out))
	a.AMQP1Sender.Send(string(out))

	// We can shortcut this: since renderTemplate returns `error`,
	// our ServeHTTP method will return a HTTP 500 instead and won't
	// attempt to write a broken template out with a HTTP 200 status.
	// (see the postscript for how renderTemplate is implemented)
	// If it doesn't return an error, things will go as planned.
	return http.StatusOK, nil
}
