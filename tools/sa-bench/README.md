# sa-bench

## Description

sa-bench works as AMQP message sender, sending collectd simulated metrics json.

## Requirement
Install [qpid-proton](https://github.com/apache/qpid-proton/tree/master/examples/go#using-the-go-packages) with go client.
See [https://github.com/apache/qpid-proton/tree/master/examples/go#using-the-go-packages](https://github.com/apache/qpid-proton/tree/master/examples/go#using-the-go-packages) for the details.

## How to build

    # need to install
    go build main.go

## Usage

    usage: ./sa-bench (options) ampq://...
    options:
      -hosts int
                Simulate hosts (default 1)
      -interval int
                Interval (sec) (default 1)
      -metrics int
                Metrics per hosts (default 1)
      -send int
                How many metrics sent (default 1)

### Example1

     # Send one json data from one host metric to amqp
     $ ./sa-bench amqp://localhost:5672/foo

### Example2

     # Simulate sending json data 3 times, each from 2 hosts with 5sec intervals, (total 2 * 3 = 6 json messages are sent)
     $ ./sa-bench -hosts 2 -interval 5 -metrics 1 -send 3 amqp://localhost:5672/foo

### Author
- Tomofumi Hayashi (s1061123)
