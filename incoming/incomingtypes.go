package incoming

//IncomingDataInterface   ...
type IncomingDataInterface interface {
	GetName() string
	SetData(data IncomingDataInterface)
	ParseInputJSON(json string) error
	GetKey() string
	GetItemKey() string
	GenerateSampleData(key string, itemkey string) IncomingDataInterface
	GenerateSampleJSON(key string, itemkey string) string
	ParseInputByte(data []byte) error
	//GenerateSamples(jsonstring string) *Interface
	SetNew(new bool)
	ISNew() bool
	TSDB
}

//TSDB  interface
type TSDB interface {
	//prometheus specifivreflect
	GetLabels() map[string]string
	GetMetricName(index int) string
	GetMetricDesc(index int) string
}

//IncomingDataType   ..
type IncomingDataType int

//COLLECTD
const (
	COLLECTD IncomingDataType = 1 << iota
)

//NewInComing   ..
func NewInComing(t IncomingDataType) IncomingDataInterface {
	switch t {
	case COLLECTD:
		return newCollectd( /*...*/ )
	}
	return nil
}

//newCollectd  -- avoid calling this . Use factory method in incoming package
func newCollectd() *Collectd {
	return new(Collectd)
}

//GenerateData  Generates sample data in source format
func GenerateData(dataItem IncomingDataInterface, key string, itemkey string) IncomingDataInterface {
	return dataItem.GenerateSampleData(key, itemkey)
}

//GenerateJSON  Generates sample data  in json format
func GenerateJSON(dataItem IncomingDataInterface, key string, itemkey string) string {
	return dataItem.GenerateSampleJSON(key, itemkey)
}

//ParseByte  parse incoming data
func ParseByte(dataItem IncomingDataInterface, data []byte) error {
	return dataItem.ParseInputByte(data)
}
