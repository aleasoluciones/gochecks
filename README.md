#gochecks
[![Build Status](https://travis-ci.org/aleasoluciones/gochecks.svg?branch=master)](https://travis-ci.org/aleasoluciones/gochecks)
[![GoDoc](https://godoc.org/github.com/aleasoluciones/gochecks?status.png)](http://godoc.org/github.com/aleasoluciones/gochecks)

gochecks package provides utilities to check services health and publish events. 
It is written fully in Go. 

It includes:
 * checks scheduler
 * various checks:
   * Tcp port
   * ICMP/Ping
   * http
   * snmp get
   * rabbitmq queue len
   * Arris C4 CMTS temp
   * JunOS devices cpu usage and temp
   * MySQL connectivity

 * Publishers:
  * [riemann](http://riemann.io/)
  * RabbitMQ / AMQP

Install
=======

```
go get github.com/aleasoluciones/gochecks
```

Development
===========
Export vars:
MYSQL_URL=mysql://<user>:<pass>@<host>/<database>
AMQP_URL=qmqp://<user>:<pass>@<host>/<vhost>

```
go test -tags integration -v ./...
```

Todo
====
 * Metric for all the checks
 * Nagios output format
