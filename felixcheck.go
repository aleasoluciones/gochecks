package felixcheck

import (
	"fmt"
	"log"
	"net"
	"time"

	//"encoding/json"
	"net/http"

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

type CheckResult struct {
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

func (ce CheckEngine) AddCheck(host, service string, period time.Duration, check CheckFunction) {
	scheduledtask.NewScheduledTask(func() {
		result, err := check()
		log.Println("Result", host, service, result, err)
	}, period, 0)
}

type CheckPublisher interface {
	PublishCheckResult(result CheckResult)
}

func (p RabbitMqPublisher) PublishCheckResult(result CheckResult) {
	// var state string
	// if result.result == true {
	// 	state = "ok"
	// } else {
	// 	state = "critical"
	// }
	// topic := fmt.Sprintf("check.%s.%s", result.checker, result.device.Id)
	// serialized, _ := json.Marshal(CheckResultMessage{result.device.Id, result.service, state})
	// p.publisher.Publish(topic, serialized)
}

type CheckFunction func() (bool, error)

func NewPingCheck(ip string) CheckFunction {
	return func() (bool, error) {
		var retRtt time.Duration = 0
		var isUp bool = false

		p := fastping.NewPinger()
		p.MaxRTT = maxPingTime
		ra, err := net.ResolveIPAddr("ip4:icmp", ip)

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

func NewTcpPortChecker(ip string, port int, conf TcpCheckerConf) CheckFunction {
	return func() (bool, error) {
		var err error
		var conn net.Conn

		for attempt := 0; attempt < conf.retries; attempt++ {
			conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), conf.timeout)
			if err == nil {
				conn.Close()
				return true, nil
			}
			time.Sleep(conf.retrytime)
		}
		return false, err
	}
}

func NewHttpChecker(url string, expectedStatusCode int) CheckFunction {
	return func() (bool, error) {
		response, err := http.Get(url)
		if err != nil {
			return false, err
		} else {
			defer response.Body.Close()
			if response.StatusCode == expectedStatusCode {
				return true, nil
			}
		}
		return false, err
	}
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

func NewSnmpChecker(ip, community string, conf SnmpCheckerConf, snmpQuerier gosnmpquerier.SyncQuerier) CheckFunction {
	return func() (bool, error) {
		_, err := snmpQuerier.Get(ip, community, []string{conf.oidToCheck}, conf.timeout, conf.retries)
		if err == nil {
			return true, nil
		} else {
			return false, err
		}
	}
}
