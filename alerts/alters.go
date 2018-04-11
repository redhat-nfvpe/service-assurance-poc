package alerts

import (
	"encoding/json"
	"log"
	"strings"
)

//Alerts  ...
type Alerts struct {
	Status       string            `json:"status"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
	StartsAt     string            `json:"startsAt,omitempty"`
	EndsAt       string            `json:"endsAt,omitempty"`
	GeneratorURL string            `json:"generatorURL"`
}

//SetName ... set unique name for alerts
func (a *Alerts) SetName() {
	values := make([]string, 0, len(a.Labels)-1)
	for k, v := range a.Labels {
		if k != "severity" {
			values = append(values, v)
		}
	}
	a.Labels["name"] = strings.Join(values, "_")
}

//Parse ...parses alerts to validate for schema
func (a *Alerts) Parse(eventJSON []byte, generatorURL string) {
	var dat []map[string]interface{}
	a.GeneratorURL = generatorURL
	if err := json.Unmarshal(eventJSON, &dat); err != nil {
		log.Println("Error parsing events for alerts.")
		log.Panic(err)
	}
	a.Annotations = make(map[string]string)
	a.Labels = make(map[string]string)
	labels := dat[0]["labels"].(map[string]interface{})
	for k, v := range labels {
		a.Labels[k] = v.(string)
		switch k {
		case "severity":
			switch v.(string) {
			case "OKAY":
				a.Status = "resolved"
			default:
				a.Status = "firing"
			}
		}
	}
	a.SetName()
	a.Labels["alertsource"] = "SMARTAGENT"
}
