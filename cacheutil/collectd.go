package cacheutil

import (
	"encoding/json"
	"fmt"

	"strconv"
)

type Collectd struct {
	Values          []float64
	Dstypes         []string
	Dsnames         []string
	Time            int64
	Interval        int
	Host            string
	Plugin          string
	Plugin_instance string
	Type            string `json:"type"`
	Type_instance   string
	new             bool
}

//ParseCollectdJSON   ...
func ParseCollectdJSON(collectdJson string) *Collectd {
	c := make([]Collectd, 1)
	var jsonBlob = []byte(collectdJson)
	err := json.Unmarshal(jsonBlob, &c)
	if err != nil {
		fmt.Println("error:", err)
	}
	c1 := c[0]
	c1.SetNew(true)
	return &c1

}

//ParseCollectdByte ....
func ParseCollectdByte(amqpCollectd []byte) *Collectd {
	c := make([]Collectd, 1)
	//var jsonBlob = []byte(collectdJson)
	err := json.Unmarshal(amqpCollectd, &c)
	if err != nil {
		fmt.Println("error:", err)
	}
	c1 := c[0]
	c1.SetNew(true)
	return &c1

}

func (c *Collectd) SetNew(new bool) {
	c.new = new
}
func (c *Collectd) ISNew() bool {
	return c.new
}

// newName converts one data source of a value list to a string representation.
func (vl *Collectd) DSName(index int) string {
	if vl.Dsnames != nil {
		return vl.Dsnames[index]
	} else if len(vl.Values) != 1 {
		return strconv.FormatInt(int64(index), 10)
	}

	return "value"
}

//generateCollectdJson   for samples
func GenerateCollectdJson(hostname string, pluginname string) string {
	return `[{
      "values":  [0.0,0.0],
      "dstypes":  ["gauge","guage"],
      "dsnames":    ["value1","value2"],
      "time":      0,
      "interval":          10,
      "host":            "hostname",
      "plugin":          "pluginname",
      "plugin_instance": "0",
      "type":            "pluginname",
      "type_instance":   "idle"
    }]`
}
