package amqp10

import (
	"testing"
)

func TestPut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	notifier := make(chan string) // Channel for messages from goroutines to main()
	var url = "amqp://10.19.110.5:5672/collectd/telemetry"
	var amqpServer *AMQPServer
	amqpServer = NewAMQPServer(url, true, 10, notifier)

	for i := 0; i < 10; i++ {
		data := <-amqpServer.notifier
		t.Logf("%s\n", data)
	}

}
