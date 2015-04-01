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

Install
=======

```
go get github.com/aleasoluciones/felixcheck
```

Todo
====
 * [riemann](http://riemann.io/) publisher
 * include metrics with each check
