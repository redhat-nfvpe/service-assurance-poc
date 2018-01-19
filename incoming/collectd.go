package incoming

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"
  "log"
)

type Collectd struct {
	Values          []float64
	Dstypes         []string
	Dsnames         []string
	Time            float64 `json:"time"`
	Interval        float64 `json:"interval"`
	Host            string  `json:"host"`
	Plugin          string  `json:"plugin"`
	Plugin_instance string  `json:"plugin_instance"`
	Type            string  `json:"type"`
	Type_instance   string
	new             bool
}
//CreateNewCollectd don't use .... use incoming.NewInComing
func CreateNewCollectd() *Collectd{
  return new(Collectd)
}

//GetName implement interface
func (c Collectd) GetName() string {
	return c.Plugin
}

//GetKey ...
func (c Collectd) GetKey() string{
	return c.Host
}
//ParseInputByte ....
func (c *Collectd)ParseInputByte(data []byte) error{
	cparse := make([]Collectd, 1)
	//var jsonBlob = []byte(collectdJson)
	err := json.Unmarshal(data, &cparse)
	if err != nil {
		log.Printf("error:%v", err)
    return err
	}
  c1 := cparse[0]
	c1.SetNew(true)
	c.SetData(&c1)
  return nil
}

//SetNew  .
func (c *Collectd) SetNew(new bool) {
	c.new = new
}

//ISNew   ..
func (c *Collectd) ISNew() bool {
	return c.new
}

//DSName newName converts one data source of a value list to a string representation.
func (c *Collectd) DSName(index int) string {
	if c.Dsnames != nil {
		return c.Dsnames[index]
	} else if len(c.Values) != 1 {
		return strconv.FormatInt(int64(index), 10)
	}
	return "value"
}

//SetData   ...
func (c *Collectd) SetData(data IncomingDataInterface) {
	if collectd, ok := data.(*Collectd); ok { // type assert on it
		if c.Host != collectd.Host {
			c.Host = collectd.Host
		}
		if c.Plugin != collectd.Plugin {
			c.Plugin = collectd.Plugin
		}
		c.Interval = collectd.Interval
		c.Values = collectd.Values
		c.Dsnames = collectd.Dsnames
		c.Dstypes = collectd.Dstypes
		c.Time = collectd.Time
		if c.Plugin_instance != collectd.Plugin_instance {
			c.Plugin_instance = collectd.Plugin_instance
		}
		if c.Type != collectd.Type {
			c.Type = collectd.Type
		}
		if c.Type_instance != collectd.Type_instance {
			c.Type_instance = collectd.Type_instance
		}
		c.SetNew(true)
	}
}

//GetLabels   ..
func (c Collectd) GetLabels() map[string]string {
	labels := map[string]string{}
	if c.Plugin_instance != "" {
		labels[c.Plugin] = c.Plugin_instance
	}
	if c.Type_instance != "" {
		if c.Plugin_instance == "" {
			labels[c.Plugin] = c.Type_instance
		} else {
			labels["type"] = c.Type_instance
		}
	}
	labels["instance"] = c.Host

	return labels
}

//GetMetricDesc   newDesc converts one data source of a value list to a Prometheus description.
func (c Collectd) GetMetricDesc(index int) string {
	help := fmt.Sprintf("Service Assurance exporter: '%s' Type: '%s' Dstype: '%s' Dsname: '%s'",
		c.Plugin, c.Type, c.Dstypes[index], c.DSName(index))
	return help

}

//GetMetricName  ..
func (c Collectd) GetMetricName(index int) string {
	var name string
	if c.Plugin == c.Type {
		name = "service_assurance_collectd_" + c.Type
	} else {
		name = "service_assurance_collectd_" + c.Plugin + "_" + c.Type
	}

	if dsname := c.DSName(index); dsname != "value" {
		name += "_" + dsname
	}

	switch c.Dstypes[index] {
	case "counter", "derive":
		name += "_total"
	}
	return name

}

func (c Collectd) GetItemKey() string {
	return c.Plugin
}

//GenerateSampleData
func (c *Collectd) GenerateSampleData(hostname string, pluginname string) IncomingDataInterface {
  collectd:=CreateNewCollectd()
	collectd.Host = hostname
	collectd.Plugin = pluginname
	collectd.Type = pluginname
	collectd.Plugin_instance = pluginname
	collectd.Dstypes = []string{"guage", "gauge"}
	collectd.Dsnames = []string{"value1", "value2"}
	collectd.Type_instance = "idle"
	collectd.Values = []float64{rand.Float64(), rand.Float64()}
	collectd.Time = float64((time.Now().UnixNano())) / 1000000
	return collectd
}

//ParseCollectdJSON   ...
func (c *Collectd) ParseInputJSON(jsonString string) error {
	collect := make([]Collectd, 1)
	var jsonBlob = []byte(jsonString)
	err := json.Unmarshal(jsonBlob, &collect)
	if err != nil {
		log.Println("error:", err)
    return err
	}
	c1 := collect[0]
	c1.SetNew(true)
	c.SetData(&c1)
  return nil

}

//generateCollectdJson   for samples
func (c Collectd) GenerateSampleJson(hostname string, pluginname string) string {
	return `[{
      "values":  [0.0,0.0],
      "dstypes":  ["gauge","guage"],
      "dsnames":    ["value11","value12"],
      "time":      0.0,
      "interval":          10.0,
      "host":            "hostname",
      "plugin":          "apluginname",
      "plugin_instance": "0",
      "type":            "pluginname",
      "type_instance":   "idle"
    }]`
}
