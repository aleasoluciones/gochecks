# gochecks

[![Build Status](https://travis-ci.org/aleasoluciones/gochecks.svg?branch=master)](https://travis-ci.org/aleasoluciones/gochecks)
[![GoDoc](https://godoc.org/github.com/aleasoluciones/gochecks?status.png)](http://godoc.org/github.com/aleasoluciones/gochecks)

gochecks package provides utilities to check services health and publish events.
It is written fully in Go.

It includes:
 * checks scheduler
 * add checks results
 * various checks:
   * Tcp port
   * ICMP/Ping
   * http
   * snmp get
   * rabbitmq queue len
   * Arris C4 CMTS temp
   * JunOS devices cpu usage and temp
   * MySQL connectivity
   * Jenkins jobs status

 * Publishers:
  * [riemann](http://riemann.io/)
  * RabbitMQ / AMQP

##Install

```
go get github.com/aleasoluciones/gochecks
```


##Sample code

Create a Checks Engine with two publisher (riemman and log)
```
checkEngine := gochecks.NewCheckEngine([]gochecks.CheckPublisher{
    gochecks.NewRiemannPublisher("127.0.0.1:5555"),
    gochecks.NewLogPublisher(),
})
```
Add a periodic (20 seconds) http check with up to three retries, tagging the result as production and adding some attributes.
```
checkEngine.AddCheck(
    gochecks.NewHttpChecker("golang", "http", "http://www.golang.org", 200).
      Attributes(map[string]string{"version": "1", "network": "google"}).
      Tags("production").
      Retry(3, 1*time.Second),
    20 * time.Second)
```

##Development

To pass the integration tests you need to execute a MySQL server, a Postgres server and a RabbitMQ Server and export the corresponding vars.

The directory dev/ offers two scripts and an environment file to achieve this easily:

```
source dev/env_develop
dev/start_gochecks_dependencies.sh
go test -tags integration -v ./...
dev/stop_gochecks_dependencies.sh
```

##Todo
 * Metric for all the checks
