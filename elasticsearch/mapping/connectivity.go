package elastic
const ConnectivityMapping = `
{
   "mappings": {
      "event": {
         "properties": {
            "labels": {
               "type": "object",
               "properties": {
                  "alertname": {
                     "type": "text",
                     "index": true
                  },
                  "instance": {
                     "type": "text",
                     "index": true
                  },
                  "connectivity": {
                     "type": "text",
                     "index": true
                  },
                  "type": {
                     "type": "text",
                     "index": true
                  },
                  "severity": {
                     "type": "text",
                     "index": true
                  },
                  "service": {
                     "type": "text",
                     "index": true
                  }
               }
            },
            "annotations": {
               "type": "object",
               "properties": {
                  "summary": {
                     "type": "text",
                     "index": false
                  },
                  "ves": {
                     "type": "object",
                     "properties": {
                        "domain": {
                           "type": "text",
                            "index": true
                        },
                        "stateChange": {
                           "type": "text",
                            "index": true
                        },
                        "eventId": {
                           "type": "text",
                           "index": false
                        },
                        "eventName": {
                           "type": "text",
                            "index": true
                        },
                        "lastEpochMicrosec": {
                           "type": "date",
                           "format": "strict_date_optional_time||epoch_millis"

                        },
                        "priority": {
                           "type": "text",
                            "index": false
                        },
                        "reportingEntityName": {
                           "type": "text",
                           "index": true
                        },
                        "sequence": {
                           "type": "integer",
                           "index": false
                        },
                        "sourceName": {
                           "type": "text",
                            "index": true
                        },
                        "startEpochMicrosec": {
                           "type": "date",
                           "format": "strict_date_optional_time||epoch_millis"

                        },
                        "version": {
                           "type": "float",
                           "index":false
                        },
                        "stateChangeFields": {
                           "type": "object",
                           "properties": {
                              "newState": {
                                 "type": "text",
                                  "index": true
                              },
                              "oldState": {
                                  "type": "text",
                                   "index": true
                              },
                              "stateChangeFieldsVersion": {
                                 "type": "float",
                                 "index": false

                              }
                           }
                        }
                     }
                  }
               }
            },
            "startsAt": {
               "type": "text"
            }
         }
       }
      }
}
`
