package felixcheck

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/aleasoluciones/goaleasoluciones/scheduledtask"
	"github.com/aleasoluciones/gosnmpquerier"
)

const (
	sysName = "1.3.6.1.2.1.1.5.0"
)

type Device struct {
	Id        string
	DevType   string
	Ip        string
	Community string
}

type Checker struct {
	Devices     []Device
	SnmpQuerier gosnmpquerier.SyncQuerier
}

func NewChecker(devices []Device) Checker {
	return Checker{Devices: devices, SnmpQuerier: gosnmpquerier.NewSyncQuerier(1, 1, 4*time.Second)}
}

func (c *Checker) Start() {
	for _, device := range c.Devices {
		if device.DevType == "bos" {
			c.checkTcpPortLoop(device, 6922)
		}
		if device.Community != "" {
			c.checkSnmpLoop(device)
		}
	}
}

func (c *Checker) checkSnmpLoop(device Device) {
	scheduledtask.NewScheduledTask(func() {
		result, err := c.SnmpQuerier.Get(device.Ip, device.Community, []string{sysName}, 1*time.Second, 1)
		if err == nil {
			log.Println("Check snmp ok", device, result)
		} else {
			log.Println("Check snmp error", device, err)
		}
	}, 20*time.Second, 0)
}

func (c *Checker) checkTcpPortLoop(device Device, port int) {
	scheduledtask.NewScheduledTask(func() {
		if ok, err := c.checkTcpPort(device, port); ok {
			log.Println("Check tcp ok", device)
		} else {
			log.Println("Check tcp error", device, err)
		}
	}, 20*time.Second, 0)
}

func (c *Checker) checkTcpPort(device Device, port int) (bool, error) {
	var err error
	var conn net.Conn

	for attempt := 0; attempt < 3; attempt++ {
		conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", device.Ip, port), 2*time.Second)
		if err == nil {
			conn.Close()
			return true, nil
		}
		time.Sleep(1 * time.Second)
	}
	return false, err
}
