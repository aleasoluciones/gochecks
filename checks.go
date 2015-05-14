package felixcheck

import (
	"fmt"
	"net"
	"strings"
	"time"

	"net/http"
	"net/url"

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
type MultiCheckFunction func() []goryman.Event

func (f CheckFunction) Tags(tags ...string) CheckFunction {
	return func() goryman.Event {
		result := f()
		result.Tags = tags
		return result
	}
}

func (f CheckFunction) Ttl(ttl float32) CheckFunction {
	return func() goryman.Event {
		result := f()
		result.Ttl = ttl
		return result
	}
}

func NewPingChecker(host, service, ip string) CheckFunction {
	return func() goryman.Event {
		var retRtt time.Duration = 0
		var result goryman.Event = goryman.Event{Host: host, Service: service, State: "critical"}

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

func NewTcpPortChecker(host, service, ip string, port int, conf TcpCheckerConf) CheckFunction {
	return func() goryman.Event {
		var err error
		var conn net.Conn

		for attempt := 0; attempt < conf.retries; attempt++ {
			conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), conf.timeout)
			if err == nil {
				conn.Close()
				return goryman.Event{Host: host, Service: service, State: "ok"}
			}
			time.Sleep(conf.retrytime)
		}
		return goryman.Event{Host: host, Service: service, State: "critical"}
	}
}

func NewHttpChecker(host, service, url string, expectedStatusCode int) CheckFunction {
	return func() goryman.Event {
		var t1 = time.Now()

		response, err := http.Get(url)
		milliseconds := float32((time.Now().Sub(t1)).Nanoseconds() / 1e6)
		result := goryman.Event{Host: host, Service: service, State: "critical", Metric: milliseconds}
		if err != nil {
			result.Description = err.Error()
		} else {
			if response.Body != nil {
				defer response.Body.Close()
			}
			if response.StatusCode == expectedStatusCode {
				result.State = "ok"
			} else {
				result.Description = fmt.Sprintf("Response %d", response.StatusCode)
			}
		}
		return result
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

func NewSnmpChecker(host, service, ip, community string, conf SnmpCheckerConf) CheckFunction {
	return func() goryman.Event {

		_, err := snmpGet(ip, community, []string{conf.oidToCheck}, conf.timeout, conf.retries)
		if err == nil {
			return goryman.Event{Host: host, Service: service, State: "ok", Description: err.Error()}
		} else {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error()}
		}
	}
}

func NewC4CMTSTempChecker(host, service, ip, community string, maxAllowedTemp int) CheckFunction {
	return func() goryman.Event {

		result, err := snmpWalk(ip, community, "1.3.6.1.4.1.4998.1.1.10.1.4.2.1.29", 2*time.Second, 1)

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
			return goryman.Event{Host: host, Service: service, State: state, Metric: float32(max)}
		} else {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error()}
		}
	}
}

func getMaxValueFromSnmpWalk(oid, ip, community string) (uint, error) {
	result, err := snmpWalk(ip, community, oid, 2*time.Second, 1)
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

func NewJuniperTempChecker(host, service, ip, community string, maxAllowedTemp uint) CheckFunction {
	return func() goryman.Event {
		max, err := getMaxValueFromSnmpWalk("1.3.6.1.4.1.2636.3.1.13.1.7", ip, community)
		if err == nil {
			var state string = "critical"
			if max < maxAllowedTemp {
				state = "ok"
			}
			return goryman.Event{Host: host, Service: service, State: state, Metric: float32(max)}
		} else {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error()}
		}
	}
}

func NewJuniperCpuChecker(host, service, ip, community string, maxAllowedTemp uint) CheckFunction {
	return func() goryman.Event {
		max, err := getMaxValueFromSnmpWalk("1.3.6.1.4.1.2636.3.1.13.1.8", ip, community)
		if err == nil {
			var state string = "critical"
			if max < maxAllowedTemp {
				state = "ok"
			}
			return goryman.Event{Host: host, Service: service, State: state, Metric: float32(max)}
		} else {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error()}
		}
	}
}

func NewRabbitMQQueueLenCheck(host, service, amqpuri, queue string, max int) CheckFunction {
	return func() goryman.Event {
		queueInfo, err := simpleamqp.NewAmqpManagement(amqpuri).QueueInfo(queue)
		if err == nil {
			var state string = "critical"
			if queueInfo.Messages < max {
				state = "ok"
			}
			return goryman.Event{Host: host, Service: service, State: state, Metric: float32(queueInfo.Messages)}
		} else {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error()}
		}
	}
}

func NewMysqlConnectionCheck(host, service, mysqluri string) CheckFunction {
	return func() goryman.Event {
		u, err := url.Parse(mysqluri)
		if err != nil {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error()}
		}

		password, _ := u.User.Password()
		hostAndPort := u.Host
		if !strings.Contains(hostAndPort, ":") {
			hostAndPort = hostAndPort + ":3306"
		}
		con, err := sql.Open("mysql", u.User.Username()+":"+password+"@"+"tcp("+hostAndPort+")"+u.Path)
		defer con.Close()
		if err != nil {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error()}
		}
		q := `select CURTIME()`
		row := con.QueryRow(q)
		var date string
		err = row.Scan(&date)
		if err != nil {
			return goryman.Event{Host: host, Service: service, State: "critical", Description: err.Error()}
		}
		return goryman.Event{Host: host, Service: service, State: "ok"}
	}
}

type ObtainMetricFunction func() float32
type CalculateStateFunction func(float32) string

func NewGenericCheck(host, service string, metricFunc ObtainMetricFunction, stateFunc CalculateStateFunction) CheckFunction {
	return func() goryman.Event {
		value := metricFunc()
		var state string = stateFunc(value)
		return goryman.Event{Host: host, Service: service, State: state, Metric: value}
	}
}
