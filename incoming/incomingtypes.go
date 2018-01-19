package incoming

type  Interface interface{
  GetName() string
  SetData(data interface{})
  ParseInputJSON(json string) error
  GenerateSampleData(key string, itemkey string)
  GenerateSampleJson( key string, itemkey string) string
  //GenerateSamples(jsonstring string) *Interface
  SetNew(new bool)
  ISNew() bool
  TSDB
}

type TSDB interface{
  //prometheus specifiv
  GetLabels()map[string]string
  GetMetricName(index int)string
  GetMetricDesc(index int) string

}
type IncomingDataType int

const (
    COLLECTD IncomingDataType = 1 << iota

)

func NewInComing(t IncomingDataType) Interface {
    switch t {
    case COLLECTD:
        return newCollectd( /*...*/ )
    }
    return nil
}
//CreateFactory
func newCollectd() *Collectd{
  return new(Collectd)
}
