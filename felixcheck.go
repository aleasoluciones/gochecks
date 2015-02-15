package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"encoding/json"
	"io/ioutil"

	"github.com/aleasoluciones/goaleasoluciones/scheduledtask"
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
			devices = append(devices, Device{
				DevType: deviceMap["dev_type"].(string),
				Id:      deviceMap["id"].(string),
				Ip:      deviceMap["ip"].(string),
				//Community: deviceMap["snmp_rw"].(string),
			})
		}
	}
	return devices, nil
}

type Checker struct {
	Devices []Device
}

func NewChecker(devices []Device) Checker {
	return Checker{Devices: devices}
}

func (c *Checker) Start() {
	for _, device := range c.Devices {
		if device.DevType == "bos" {
			c.checkTcpPortLoop(device, 6922)
		}
	}
}

func (c *Checker) checkTcpPortLoop(device Device, port int) {
	log.Println("Start check bos", device)
	scheduledtask.NewScheduledTask(func() {
		if ok, err := c.checkTcpPort(device, port); ok {
			log.Println("Check ok", device)
		} else {
			log.Println("Check error", device, err)
		}
	}, 20*time.Second, 0)
}

func (c *Checker) checkTcpPort(device Device, port int) (bool, error) {
	var err error
	var conn net.Conn

	for attempt := 0; attempt < 3; attempt++ {
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", device.Ip, port))
		if err == nil {
			conn.Close()
			return true, nil
		}
		time.Sleep(1 * time.Second)
	}
	return false, err
}

func main() {
	devices, err := devicesFromInventory(os.Getenv("INVENTORY_FILE"))
	if err != nil {
		log.Panic(err)
	}
	log.Println("devices", devices)
	checker := NewChecker(devices)
	checker.Start()

	for {
		time.Sleep(2 * time.Second)
	}
}
