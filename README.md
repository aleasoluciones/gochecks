
#gochecks
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
  * RabbitMQ / AMQP

##Install

```
go get github.com/aleasoluciones/gochecks
```


##Sample code

Create a Checks Engine with one log publisher
```
checkEngine := gochecks.NewCheckEngine([]gochecks.CheckPublisher{
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

To pass the integration tests you need to execute a MySQL server, a Postgres server and a RabbitMQ Server and export the corresponding vars:

MYSQL_URL=mysql://<user>:<pass>@<host>/<database>
POSTGRES_URL=postgres://<user>:<pass>@<host>/<satabase>
AMQP_URL=amqp://<user>:<pass>@<host>/<vhost>

Example:
```
docker run -p 3306:3306 -e MYSQL_ROOT_PASSWORD=rootpass -d mysql
export MYSQL_URL=mysql://root:rootpass@localhost/mysql

docker run -p 5432:5432 -e POSTGRES_PASSWORD=mysecretpassword -d postgres:9.5
export POSTGRES_URL=postgres://postgres:mysecretpassword@localhost/postgres?sslmode=disable

docker run -p 5672:5672 -d rabbitmq
export AMQP_URL=amqp://guest:guest@127.0.0.1:5672/

go test -tags integration -v ./...
```

##Todo
 * Metric for all the checks
