package felixcheck

import (
	"fmt"
	"net"
	"strings"
	"time"

	"net/http"
	"net/url"

	"github.com/aleasoluciones/gosnmpquerier"
	"github.com/aleasoluciones/simpleamqp"
	"github.com/bigdatadev/goryman"
	"github.com/tatsushid/go-fastping"
)

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

const (
	sysName     = "1.3.6.1.2.1.1.5.0"
	maxPingTime = 4 * time.Second
)

type CheckFunction func() goryman.Event

func NewPingChecker(host, service, ip string, tags ...string) CheckFunction {
	return func() goryman.Event {
		var retRtt time.Duration = 0
		var result goryman.Event = goryman.Event{Host: host, Service: service, State: "critical", Tags: tags}

		p := fastping.NewPinger()
		p.MaxRTT = maxPingTime
		ra, err := net.ResolveIPAddr("ip4:icmp", ip)
		if err != nil {
			result.Description = err.Error()
		}

		p.AddIPAddr(ra)
		p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
			result.State = "ok"
			result.Metric = float32(retRtt.Nanoseconds() / 1e6)
		}

		err = p.Run()
		if err != nil {
			result.Description = err.Error()
		}
		return result
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

func NewTcpPortChecker(host, service, ip string, port int, conf TcpCheckerConf, tags ...string) CheckFunction {
	return func() goryman.Event {
		var err error
		var conn net.Conn

		for attempt := 0; attempt < conf.retries; attempt++ {
			conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), conf.timeout)
			if err == nil {
				conn.Close()
				return goryman.Event{Host: host, Service: service, State: "ok", Tags: tags}
			}
			time.Sleep(conf.retrytime)
		}
		return goryman.Event{Host: host, Service: service, State: "critical", Tags: tags}
	}
}

func NewHttpChecker(host, service, url string, expectedStatusCode int, tags ...string) CheckFunction {
	return func() goryman.Event {
		var t1 = time.Now()
		response, err := http.Get(url)
		milliseconds := float32((time.Now().Sub(t1)).Nanoseconds() / 1e6)
		if err != nil {
			return goryman.Event{Host: host, Service: service, State: "critical", Metric: milliseconds, Description: err.Error(), Tags: tags}
		} else {
			defer response.Body.Close()
			if response.StatusCode == expectedStatusCode {
				return goryman.Event{Host: host, Service: service, State: "ok", Metric: milliseconds, Tags: tags}
			}
		}
		return goryman.Event{Host: host, Service: service, State: "critical", Metric: milliseconds, Description: err.Error(), Tags: tags}
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

func NewSnmpChecker(host, service, ip, community string, conf SnmpCheckerConf, snmpQuerier gosnmpquerier.SyncQuerier, tags ...string) CheckFunction {
	return func() goryman.Event {
		_, err := snmpQuerier.Get(ip, community, []string{conf.oidToCheck}, conf.timeout, conf.retries)
		if err == nil {
			return goryman.Event{Host: host, Service: service, State: "ok", Description: err.Error(), Tags: tags}
		} else {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error(), Tags: tags}
		}
	}
}

func NewC4CMTSTempChecker(host, service, ip, community string, maxAllowedTemp int, snmpQuerier gosnmpquerier.SyncQuerier, tags ...string) CheckFunction {
	return func() goryman.Event {
		result, err := snmpQuerier.Walk(ip, community, "1.3.6.1.4.1.4998.1.1.10.1.4.2.1.29", 2*time.Second, 1)

		if err == nil {
			max := 0
			for _, r := range result {
				if r.Value.(int) != 999 && r.Value.(int) > max {
					max = r.Value.(int)
				}
			}
			var state string = "critical"
			if max < maxAllowedTemp {
				state = "ok"
			}
			return goryman.Event{Host: host, Service: service, State: state, Metric: float32(max), Tags: tags}
		} else {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error(), Tags: tags}
		}
	}
}

func getMaxValueFromSnmpWalk(oid, ip, community string, snmpQuerier gosnmpquerier.SyncQuerier) (uint, error) {
	result, err := snmpQuerier.Walk(ip, community, oid, 2*time.Second, 1)
	if err == nil {
		max := uint(0)
		for _, r := range result {
			if r.Value.(uint) > max {
				max = r.Value.(uint)
			}
		}
		return max, nil
	} else {
		return 0, err
	}
}

func NewJuniperTempChecker(host, service, ip, community string, maxAllowedTemp uint, snmpQuerier gosnmpquerier.SyncQuerier, tags ...string) CheckFunction {
	return func() goryman.Event {
		max, err := getMaxValueFromSnmpWalk("1.3.6.1.4.1.2636.3.1.13.1.7", ip, community, snmpQuerier)
		if err == nil {
			var state string = "critical"
			if max < maxAllowedTemp {
				state = "ok"
			}
			return goryman.Event{Host: host, Service: service, State: state, Metric: float32(max), Tags: tags}
		} else {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error(), Tags: tags}
		}
	}
}

func NewJuniperCpuChecker(host, service, ip, community string, maxAllowedTemp uint, snmpQuerier gosnmpquerier.SyncQuerier, tags ...string) CheckFunction {
	return func() goryman.Event {
		max, err := getMaxValueFromSnmpWalk("1.3.6.1.4.1.2636.3.1.13.1.8", ip, community, snmpQuerier)
		if err == nil {
			var state string = "critical"
			if max < maxAllowedTemp {
				state = "ok"
			}
			return goryman.Event{Host: host, Service: service, State: state, Metric: float32(max), Tags: tags}
		} else {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error(), Tags: tags}
		}
	}
}

func NewRabbitMQQueueLenCheck(host, service, amqpuri, queue string, max int, tags ...string) CheckFunction {
	return func() goryman.Event {
		queueInfo, err := simpleamqp.NewAmqpManagement(amqpuri).QueueInfo(queue)
		if err == nil {
			var state string = "critical"
			if queueInfo.Messages < max {
				state = "ok"
			}
			return goryman.Event{Host: host, Service: service, State: state, Metric: float32(queueInfo.Messages), Tags: tags}
		} else {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error(), Tags: tags}
		}
	}
}

func NewMysqlConnectionCheck(host, service, mysqluri string, tags ...string) CheckFunction {
	return func() goryman.Event {
		u, err := url.Parse(mysqluri)
		if err != nil {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error(), Tags: tags}
		}

		password, _ := u.User.Password()
		hostAndPort := u.Host
		if !strings.Contains(hostAndPort, ":") {
			hostAndPort = hostAndPort + ":3306"
		}
		con, err := sql.Open("mysql", u.User.Username()+":"+password+"@"+"tcp("+hostAndPort+")"+u.Path)
		defer con.Close()
		if err != nil {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error(), Tags: tags}
		}
		q := `select CURTIME()`
		row := con.QueryRow(q)
		var date string
		err = row.Scan(&date)
		if err != nil {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error(), Tags: tags}
		}
		return goryman.Event{Host: host, Service: service, State: "ok", Tags: tags}
	}
}
