#felixcheck
[![Build Status](https://travis-ci.org/aleasoluciones/felixcheck.svg?branch=master)](https://travis-ci.org/aleasoluciones/felixcheck)
[![GoDoc](https://godoc.org/github.com/aleasoluciones/felixcheck?status.png)](http://godoc.org/github.com/aleasoluciones/felixcheck)
[![Coverage Status](https://coveralls.io/repos/aleasoluciones/felixcheck/badge.svg)](https://coveralls.io/r/aleasoluciones/felixcheck)

felixcheck package provides utilities to check services health and publish events. 
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
go get github.com/aleasoluciones/felixcheck
```

Todo
====
 * Metric for all the checks
 * Nagios output format