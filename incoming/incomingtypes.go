package incoming

type  IncomingDataInterface interface{
  GetName() string
  SetData(data IncomingDataInterface)
  ParseInputJSON(json string) error
  GetKey()string
  GetItemKey()string
  GenerateSampleData(key string, itemkey string) IncomingDataInterface
  GenerateSampleJson( key string, itemkey string) string
  ParseInputByte(data []byte) error
  //GenerateSamples(jsonstring string) *Interface
  SetNew(new bool)
  ISNew() bool
  TSDB
}


//TSDB  interface
type TSDB interface{
  //prometheus specifivreflect
  GetLabels()map[string]string
  GetMetricName(index int)string
  GetMetricDesc(index int) string

}
type IncomingDataType int

const (
    COLLECTD IncomingDataType = 1 << iota

)

func NewInComing(t IncomingDataType) IncomingDataInterface {
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

func GenerateData(dataItem IncomingDataInterface,key string, itemkey string) IncomingDataInterface{
  return dataItem.GenerateSampleData(key ,itemkey)
}
func GenerateJson(dataItem IncomingDataInterface, key string, itemkey string) string{
  return dataItem.GenerateSampleJson(key,itemkey)
}
func ParseByte(dataItem IncomingDataInterface,data []byte) error {
  return dataItem.ParseInputByte(data)
}
