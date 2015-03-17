package felixcheck

import (
	"fmt"
	"net"
	"time"

	"net/http"

	"github.com/aleasoluciones/gosnmpquerier"
	"github.com/tatsushid/go-fastping"
)

const (
	sysName     = "1.3.6.1.2.1.1.5.0"
	maxPingTime = 4 * time.Second
)

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
