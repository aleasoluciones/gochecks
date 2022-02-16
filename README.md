# gochecks

[![Build Status](https://travis-ci.com/aleasoluciones/gochecks.svg?branch=master)](https://travis-ci.com/aleasoluciones/gochecks)
[![GoDoc](https://godoc.org/github.com/aleasoluciones/gochecks?status.png)](http://godoc.org/github.com/aleasoluciones/gochecks)
[![License](https://img.shields.io/github/license/aleasoluciones/http2amqp)](https://github.com/aleasoluciones/http2amqp/blob/master/LICENSE)

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
   * RabbitMQ / AMQP

## Install

```
go get github.com/aleasoluciones/gochecks
```


## Sample code

Create a Checks Engine with two publisher (rabbit and log)
```
checkEngine := gochecks.NewCheckEngine([]gochecks.CheckPublisher{
    gochecks.RabbitMqPublisher("amqp://localhost", "events"),
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

## Development

To pass the integration tests you need to execute a MySQL server, a Postgres server and a RabbitMQ Server and export the corresponding vars.

The directory dev/ offers two scripts and an environment file to achieve this easily:

```
source dev/env_develop
dev/start_gochecks_dependencies.sh
go test -tags integration -v ./...
dev/stop_gochecks_dependencies.sh
```

## Todo
 * Metric for all the checks
