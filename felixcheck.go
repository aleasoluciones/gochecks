package felixcheck

import (
	"fmt"
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

type CheckResult struct {
	device  Device
	checker Checker
	result  bool
	err     error
}

type CheckEngine struct {
	checkPublisher CheckPublisher
	results        chan CheckResult
}

func NewCheckEngine(checkPublisher CheckPublisher) CheckEngine {
	checkEngine := CheckEngine{checkPublisher, make(chan CheckResult)}
	go func() {
		for result := range checkEngine.results {
			checkEngine.checkPublisher.PublishCheckResult(result.device, result.checker, result.result, result.err)
		}
	}()
	return checkEngine
}

func (ce CheckEngine) AddCheck(device Device, c Checker, period time.Duration) {
	scheduledtask.NewScheduledTask(func() {
		result, err := c.Check(device)
		ce.results <- CheckResult{device, c, result, err}

	}, period, 0)
}

type CheckPublisher interface {
	PublishCheckResult(device Device, checker Checker, result bool, err error)
}

type Checker interface {
	Check(device Device) (bool, error)
}

type TcpCheckerConf struct {
	retries   int
	timeout   time.Duration
	retrytime time.Duration
}

var DefaultTcpCheckConf = TcpCheckerConf{
	retries:   3,
	timeout:   2 * time.Second,
	retrytime: 1 * time.Second,
}

type TcpPortChecker struct {
	port int
	conf TcpCheckerConf
}

func NewTcpPortChecker(port int, conf TcpCheckerConf) TcpPortChecker {
	return TcpPortChecker{port: port, conf: conf}
}

func (c TcpPortChecker) Check(device Device) (bool, error) {
	var err error
	var conn net.Conn

	for attempt := 0; attempt < c.conf.retries; attempt++ {
		conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", device.Ip, c.port), c.conf.timeout)
		if err == nil {
			conn.Close()
			return true, nil
		}
		time.Sleep(c.conf.retrytime)
	}
	return false, err
}

type SnmpCheckerConf struct {
	retries    int
	timeout    time.Duration
	oidToCheck string
}

var DefaultSnmpCheckConf = SnmpCheckerConf{
	retries:    1,
	timeout:    1 * time.Second,
	oidToCheck: sysName,
}

type SnmpChecker struct {
	SnmpQuerier gosnmpquerier.SyncQuerier
	conf        SnmpCheckerConf
}

func NewSnmpChecker(conf SnmpCheckerConf) SnmpChecker {
	return SnmpChecker{SnmpQuerier: gosnmpquerier.NewSyncQuerier(1, 1, 4*time.Second), conf: conf}
}

func (c SnmpChecker) Check(device Device) (bool, error) {
	_, err := c.SnmpQuerier.Get(device.Ip, device.Community, []string{c.conf.oidToCheck}, c.conf.timeout, c.conf.retries)
	if err == nil {
		return true, nil
	} else {
		return false, err
	}

}
