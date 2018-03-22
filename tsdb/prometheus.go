package tsdb

import (
	"fmt"
	"regexp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redhat-nfvpe/service-assurance-poc/incoming"
)

var (
	metricNameRe = regexp.MustCompile("[^a-zA-Z0-9_:]")
)


//NewCollectdMetric converts one data source of a value list to a Prometheus metric.
func NewCollectdMetric(collectd incoming.Collectd, index int) (prometheus.Metric, error) {
	var value float64
	var valueType prometheus.ValueType

	switch collectd.Dstypes[index] {
	case "gauge":
		value = float64(collectd.Values[index])
		valueType = prometheus.GaugeValue
	case "derive", "counter":
		value = float64(collectd.Values[index])
		valueType = prometheus.CounterValue
	default:
		return nil, fmt.Errorf("unknowdsnamen value type: %s", collectd.Dstypes[index])
	}
	labels := collectd.GetLabels()
	plabels := prometheus.Labels{}
	for key, value := range labels {
		plabels[key] = value
	}

	help := fmt.Sprintf("Service Assurance exporter: '%s' Type: '%s' Dstype: '%s' Dsname: '%s'",
		collectd.Plugin, collectd.Type, collectd.Dstypes[index], collectd.DSName(index))
	metricName := metricNameRe.ReplaceAllString(collectd.GetMetricName(index), "_")
	desc := prometheus.NewDesc(metricName, help, []string{}, plabels)

	return prometheus.NewConstMetric(desc, valueType, value)
}
