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

func main() {
	devices, err := devicesFromInventory(os.Getenv("INVENTORY_FILE"))
	if err != nil {
		log.Panic(err)
	}
	checker := felixcheck.NewChecker(devices)
	checker.Start()

	for {
		time.Sleep(2 * time.Second)
	}
}
