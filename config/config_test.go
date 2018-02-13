package saconfig

import (
	"log"
	"testing"
)

func TestMetricConfig(t *testing.T) {
	var configuration MetricConfiguration
	configuration = LoadMetricConfig("config.sa.metrics.sample.json")
	if len(configuration.AMQP1MetricURL) == 0 {
		t.Error("Empty configuration generated")
	}

}

func TestEventConfig(t *testing.T) {
	var configuration EventConfiguration
	configuration = LoadEventConfig("config.sa.events.sample.json")
	log.Printf("%v\n", configuration)
	if len(configuration.AMQP1EventURL) == 0 {
		t.Error("Empty configuration generated")
	}
}
