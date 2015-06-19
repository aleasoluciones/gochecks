package main

import (
	"time"

	"github.com/aleasoluciones/gochecks"
)

func main() {
	checkEngine := gochecks.NewCheckEngine([]gochecks.CheckPublisher{
		gochecks.NewRiemannPublisher("127.0.0.1:5555"),
		gochecks.NewLogPublisher(),
	})
	period := 5 * time.Second
	googleCheck := gochecks.NewGenericHttpChecker(
		"google", "http",
		"http://www.google.com",
		gochecks.BodyGreaterThan(10000)).Tags("production", "web").TTL(50)
	checkEngine.AddCheck(googleCheck, period)
	checkEngine.AddCheck(
		gochecks.NewHttpChecker("golang", "http", "http://www.golang.org", 200).
			Attributes(map[string]string{"version": "1", "network": "google"}).
			Tags("production").
			Retry(3, 1*time.Second),
		period)

	checkEngine.AddCheck(
		gochecks.NewSnmpChecker("localhost", "snmp", "127.0.0.1", "public", gochecks.DefaultSnmpCheckConf),
		period)
	checkEngine.AddCheck(
		gochecks.NewPingChecker("nonexistinghost", "ping", "172.16.5.5").Retry(3, 1*time.Second),
		period)

	for {
		time.Sleep(2 * time.Second)
	}
}
