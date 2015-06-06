package main

import (
	"time"

	"github.com/aleasoluciones/felixcheck"
)

func main() {
	checkEngine := felixcheck.NewCheckEngine(felixcheck.NewRiemannPublisher("127.0.0.1:5555"))
	period := 5 * time.Second
	googleCheck := felixcheck.NewGenericHttpChecker(
		"google", "http",
		"http://www.google.com",
		felixcheck.BodyGreaterThan(10000)).Tags("production", "web").Ttl(50)
	checkEngine.AddCheck(googleCheck, period)
	checkEngine.AddCheck(
		felixcheck.NewHttpChecker("golang", "http", "http://www.golang.org", 200).Attributes(map[string]string{"version": "1", "network": "google"}).Tags("production", "web").Ttl(50),
		period)

	checkEngine.AddCheck(
		felixcheck.NewSnmpChecker("localhost", "snmp", "127.0.0.1", "public", felixcheck.DefaultSnmpCheckConf),
		period)
	checkEngine.AddCheck(
		felixcheck.NewPingChecker("nonexistinghost", "ping", "172.16.5.5").Retry(3, 1*time.Second),
		period)

	for {
		time.Sleep(2 * time.Second)
	}
}
