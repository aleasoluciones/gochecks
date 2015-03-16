package main

import (
	"flag"
	"log"
	"os"
	"time"

	"encoding/json"
	"io/ioutil"

	"github.com/aleasoluciones/felixcheck"
	"github.com/aleasoluciones/gosnmpquerier"
)

type Device struct {
	Id        string
	DevType   string
	Ip        string
	Community string
}

func devicesFromInventory(inventoryPath string) ([]Device, error) {
	content, err := ioutil.ReadFile(inventoryPath)
	if err != nil {
		return []Device{}, err
	}
	var inventory interface{}
	err = json.Unmarshal(content, &inventory)
	if err != nil {
		return []Device{}, err
	}

	devices := []Device{}

	for _, device := range inventory.([]interface{}) {
		deviceMap := device.(map[string]interface{})
		status := deviceMap["status"].(string)
		if status == "Activo" {
			device := Device{
				DevType: deviceMap["dev_type"].(string),
				Id:      deviceMap["id"].(string),
				Ip:      deviceMap["ip"].(string),
			}
			community, ok := deviceMap["snmp_rw"]
			if ok {
				device.Community = community.(string)
			}
			devices = append(devices, device)
		}
	}
	return devices, nil
}

type ConsoleLogPublisher struct {
}

func main() {
	devices, err := devicesFromInventory(os.Getenv("INVENTORY_FILE"))
	if err != nil {
		log.Panic(err)
	}

	var amqpuri, exchange string

	flag.StringVar(&amqpuri, "amqpuri", "amqp://guest:guest@localhost/", "AMQP connection uri")
	flag.StringVar(&exchange, "exchange", "events", "AMQP exchange")
	flag.Parse()

	publisher := felixcheck.NewRabbitMqPublisher(amqpuri, exchange)

	checkEngine := felixcheck.NewCheckEngine(publisher)
	snmpQuerier := gosnmpquerier.NewSyncQuerier(1, 1, 4*time.Second)

	for _, device := range devices {
		if device.DevType == "bos" {
			checkEngine.AddCheck(device.Id, "tcport", 20*time.Second, felixcheck.NewTcpPortChecker(device.Ip, 6922, felixcheck.DefaultTcpCheckConf))
			checkEngine.AddCheck(device.Id, "ping", 20*time.Second, felixcheck.NewPingCheck(device.Ip))
		} else {
			checkEngine.AddCheck(device.Id, "ping", 20*time.Second, felixcheck.NewPingCheck(device.Ip))
		}

		if device.Community != "" {
			checkEngine.AddCheck(device.Id, "snmp", 20*time.Second,
				felixcheck.NewSnmpChecker(device.Ip, device.Community, felixcheck.DefaultSnmpCheckConf, snmpQuerier))
		}
	}
	checkEngine.AddCheck("golang", "http", 30*time.Second, felixcheck.NewHttpChecker("http://golang.org/", 200))

	for {
		time.Sleep(2 * time.Second)
	}
}
