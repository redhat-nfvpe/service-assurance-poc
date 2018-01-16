package amqp
import (
  "fmt"
  "testing"
  "qpid.apache.org/amqp"


)

func TestPut(t *testing.T){
  messages := make(chan amqp.Message) // Channel for messages from goroutines to main()
  defer close(messages)
  var url="http://localhost:5672"
  go func(){
    AMQP(url,false,messages)
  }()

  for {
    data:=<-messages
    fmt.Printf("%v\n", data.Body())
  	}

}
