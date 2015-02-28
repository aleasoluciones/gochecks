package main

import (
	"log"
	"os"
	"time"

	"encoding/json"
	"io/ioutil"

	"github.com/aleasoluciones/felixcheck"
)

func devicesFromInventory(inventoryPath string) ([]felixcheck.Device, error) {
	content, err := ioutil.ReadFile(inventoryPath)
	if err != nil {
		return []felixcheck.Device{}, err
	}
	var inventory interface{}
	err = json.Unmarshal(content, &inventory)
	if err != nil {
		return []felixcheck.Device{}, err
	}

	devices := []felixcheck.Device{}

	for _, device := range inventory.([]interface{}) {
		deviceMap := device.(map[string]interface{})
		status := deviceMap["status"].(string)
		if status == "Activo" {
			device := felixcheck.Device{
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

func (c ConsoleLogPublisher) PublishCheckResult(device felixcheck.Device, checker felixcheck.Checker, result bool, err error) {
	log.Println("Check ", device, checker, result, err)
}

func main() {
	devices, err := devicesFromInventory(os.Getenv("INVENTORY_FILE"))
	if err != nil {
		log.Panic(err)
	}

	checkEngine := felixcheck.NewCheckEngine(ConsoleLogPublisher{})
	snmpChecker := felixcheck.NewSnmpChecker()
	tcpPortChecker := felixcheck.NewTcpPortChecker(6922, felixcheck.DefaultTcpCheckConf)

	for _, device := range devices {
		if device.DevType == "bos" {
			checkEngine.AddCheck(device, tcpPortChecker, 20*time.Second)
		}
		if device.Community != "" {
			checkEngine.AddCheck(device, snmpChecker, 20*time.Second)
		}
	}

	for {
		time.Sleep(2 * time.Second)
	}
}
