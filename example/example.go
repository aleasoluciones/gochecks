package main

import (
	"time"

	"github.com/aleasoluciones/felixcheck"
)

func main() {
	checkEngine := felixcheck.NewCheckEngine(felixcheck.NewRiemannPublisher("127.0.0.1:5555"))
	period := 5 * time.Second
	checkEngine.AddCheck(
		felixcheck.NewHttpChecker("golang", "http", "http://www.golang.org", 200).Tags("production", "web").Ttl(50),
		period)

	checkEngine.AddCheck(
		felixcheck.NewSnmpChecker("localhost", "snmp", "127.0.0.1", "public", felixcheck.DefaultSnmpCheckConf),
		period)

	for {
		time.Sleep(2 * time.Second)
	}
}
