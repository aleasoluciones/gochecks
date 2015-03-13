package felixcheck

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"encoding/json"

	"github.com/aleasoluciones/goaleasoluciones/scheduledtask"
	"github.com/aleasoluciones/gosnmpquerier"
	"github.com/aleasoluciones/simpleamqp"
	"github.com/tatsushid/go-fastping"
)

const (
	sysName     = "1.3.6.1.2.1.1.5.0"
	maxPingTime = 4 * time.Second
)

type CheckResultMessage struct {
	Host    string `json:"host"`
	Service string `json:"service"`
	State   string `json:"state"`
}

type RabbitMqPublisher struct {
	publisher *simpleamqp.AmqpPublisher
}

func NewRabbitMqPublisher(amqpuri, exchange string) RabbitMqPublisher {
	p := RabbitMqPublisher{simpleamqp.NewAmqpPublisher(amqpuri, exchange)}
	return p
}

func (p RabbitMqPublisher) PublishCheckResult(result CheckResult) {
	var state string
	if result.result == true {
		state = "ok"
	} else {
		state = "critical"
	}
	topic := fmt.Sprintf("check.%s.%s", result.checker, result.device.Id)
	serialized, _ := json.Marshal(CheckResultMessage{result.device.Id, result.service, state})
	p.publisher.Publish(topic, serialized)
}

type Device struct {
	Id        string
	DevType   string
	Ip        string
	Community string
}

type CheckResult struct {
	device  Device
	checker Checker
	service string
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
			checkEngine.checkPublisher.PublishCheckResult(result)
		}
	}()
	return checkEngine
}

func (ce CheckEngine) AddCheck(device Device, c Checker, service string, period time.Duration) {
	scheduledtask.NewScheduledTask(func() {
		result, err := c.Check(device)
		ce.results <- CheckResult{device, c, service, result, err}

	}, period, 0)
}

type CheckPublisher interface {
	PublishCheckResult(result CheckResult)
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

func (c TcpPortChecker) String() string {
	return fmt.Sprintf("TcpPortChecker %d", c.port)
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

type HttpChecker struct {
	url string
}

func NewHttpChecker(url string) HttpChecker {
	return HttpChecker{url: url}
}

func (c HttpChecker) String() string {
	return fmt.Sprintf("HttpChecker %s", c.url)
}

func (c HttpChecker) Check(device Device) (bool, error) {
	response, err := http.Get(c.url)
	if err != nil {
		return false, err
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", string(contents))
	}
	return true, err
}

type ICMPChecker struct {
}

func NewICMPChecker() ICMPChecker {
	return ICMPChecker{}
}

func (c ICMPChecker) String() string {
	return "ICMPChecker"
}

func (c ICMPChecker) Check(device Device) (bool, error) {
	var retRtt time.Duration = 0
	var isUp bool = false

	p := fastping.NewPinger()
	p.MaxRTT = maxPingTime
	ra, err := net.ResolveIPAddr("ip4:icmp", device.Ip)

	if err != nil {
		return false, err
	}

	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		isUp = true
		retRtt = rtt
	}

	err = p.Run()
	if err != nil {
		return false, err
	}

	return isUp, nil
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

func (c SnmpChecker) String() string {
	return "SNMPChecker"
}

func (c SnmpChecker) Check(device Device) (bool, error) {
	_, err := c.SnmpQuerier.Get(device.Ip, device.Community, []string{c.conf.oidToCheck}, c.conf.timeout, c.conf.retries)
	if err == nil {
		return true, nil
	} else {
		return false, err
	}

}
