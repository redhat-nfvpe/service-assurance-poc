package incoming

import (
	"testing"
  "strconv"
)

func TestCollected(t *testing.T) {
	c := CreateNewCollectd()
	jsonString := c.GenerateSampleJson("hostname", "plugi_name")
	if len(jsonString) == 0 {
		t.Error("Empty sample string generated")
	}

	c1 := CreateNewCollectd()
	if len(c1.Plugin) != 0 {
		t.Error("Collectd data  is not empty.")
	}
	//2
	c1.GenerateSampleData("hostname", "plugi_name")
	if len(c1.Plugin) == 0 {
		t.Errorf("Collectd data was not populated by GenrateSampleData %v", c1)
	}
	c1 = CreateNewCollectd()
	c1.ParseInputJSON(jsonString)
	if len(c1.Plugin) == 0 {
		t.Errorf("Collectd data was not populated by ParsestrconvInputJSON %#v", c1)
	}
	//check DSName method
	for index := range c1.Values {
		dsname := c1.DSName(index)
		if len(dsname) == 0 {
			t.Errorf("Collectd DSName is empty %#v", dsname)
		}
	}
  //pass all DSname
  c1.Dsnames=nil
  dsname:=c1.DSName(0)
  if dsname != strconv.FormatInt(int64(0), 10) {
    t.Errorf("Collectd DSName is not eq to value %s", strconv.FormatInt(int64(0), 10))
  }
  c1.Values=[]float64{1}
  dsname=c1.DSName(0)
  if dsname != "value" {
    t.Errorf("Collectd DSName is not eq to value %s", dsname)
  }


	c1 =CreateNewCollectd()
	c1.ParseCollectdByte([]byte(jsonString))
	if len(c1.Plugin) == 0 {
		t.Errorf("Collectd data was not populated by ParseCollectdByte %#v", c1)
	}
  errors:=c1.ParseCollectdByte([]byte("error string"))
  if errors==nil {
    t.Errorf("Excepted error got nil%v",errors)
  }
}

func TestCollectedMetrics(t *testing.T) {
	c1 :=CreateNewCollectd()
  c:=CreateNewCollectd()
	jsonString := c.GenerateSampleJson("hostname", "plugi_name")
	if len(jsonString) == 0 {
		t.Error("Empty sample string generated")
	}
	c1.ParseInputJSON(jsonString)
	if len(c1.Plugin) == 0 {
		t.Errorf("Collectd data was not populated by ParseInputJSON %#v", c1)
	}
  errors:=c1.ParseInputJSON("Error Json")
  if errors==nil {
    t.Errorf("Excepted error got nil%v",errors)
  }
	labels := c1.GetLabels()
	if len(labels) <2 {
		t.Errorf("Labels not populated by GetLabels %#v", c1)
	}
  name:=c1.GetName()
  if len(name) == 0 {
		t.Errorf("name not populated by GetName %#v", c1)
	}
  metricDesc:=c1.GetMetricDesc(0)
  if len(metricDesc) == 0 {
		t.Errorf("metricDesc not populated by GetMetricDesc %#v", c1)
	}
  metricName:=c1.GetMetricName(0)
  if len(metricName) == 0 {
		t.Errorf("metricName not populated by GetMetricName %#v", c1)
	}



}
