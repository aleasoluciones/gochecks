felixcheck
==========

felixcheck package provides utilities to check network services health. It is written fully in Go. 
It includes:
 * checks scheduler
 * various checks:
   * Tcp port
   * ICMP/Ping
   * http
   * snmp get
   * rabbitmq queue len
   * Arris C4 CMTS temp

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
 * include tags with each check result
