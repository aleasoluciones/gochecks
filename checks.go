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

type CheckFunction func() (bool, error, float32)

func NewPingChecker(ip string) CheckFunction {
	return func() (bool, error, float32) {
		var retRtt time.Duration = 0
		var isUp bool = false

		p := fastping.NewPinger()
		p.MaxRTT = maxPingTime
		ra, err := net.ResolveIPAddr("ip4:icmp", ip)

		if err != nil {
			return false, err, 0
		}

		p.AddIPAddr(ra)
		p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
			isUp = true
			retRtt = rtt
		}

		err = p.Run()
		if err != nil {
			return false, err, 0
		}

		return isUp, nil, float32(retRtt.Nanoseconds() / 1e6)
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
	return func() (bool, error, float32) {
		var err error
		var conn net.Conn

		for attempt := 0; attempt < conf.retries; attempt++ {
			conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), conf.timeout)
			if err == nil {
				conn.Close()
				return true, nil, 0
			}
			time.Sleep(conf.retrytime)
		}
		return false, err, 0
	}
}

func NewHttpChecker(url string, expectedStatusCode int) CheckFunction {
	return func() (bool, error, float32) {
		var t1 = time.Now()
		response, err := http.Get(url)
		milliseconds := float32((time.Now().Sub(t1)).Nanoseconds() / 1e6)
		if err != nil {
			return false, err, milliseconds
		} else {
			defer response.Body.Close()
			if response.StatusCode == expectedStatusCode {
				return true, nil, milliseconds
			}
		}
		return false, err, milliseconds
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
	return func() (bool, error, float32) {
		_, err := snmpQuerier.Get(ip, community, []string{conf.oidToCheck}, conf.timeout, conf.retries)
		if err == nil {
			return true, nil, 0
		} else {
			return false, err, 0
		}
	}
}

func NewC4CMTSTempChecker(ip, community string, maxAllowedTemp int, snmpQuerier gosnmpquerier.SyncQuerier) CheckFunction {
	return func() (bool, error, float32) {
		result, err := snmpQuerier.Walk(ip, community, "1.3.6.1.4.1.4998.1.1.10.1.4.2.1.29", 2*time.Second, 1)

		if err == nil {
			max := 0
			for _, r := range result {
				if r.Value.(int) != 999 && r.Value.(int) > max {
					max = r.Value.(int)
				}
			}
			return max < maxAllowedTemp, nil, float32(max)
		} else {
			return false, err, 0
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

func NewJuniperTempChecker(ip, community string, maxAllowedTemp uint, snmpQuerier gosnmpquerier.SyncQuerier) CheckFunction {
	return func() (bool, error, float32) {
		max, err := getMaxValueFromSnmpWalk("1.3.6.1.4.1.2636.3.1.13.1.7", ip, community, snmpQuerier)
		if err == nil {
			return max < maxAllowedTemp, nil, float32(max)
		} else {
			return false, err, 0
		}
	}
}

func NewJuniperCpuChecker(ip, community string, maxAllowedTemp uint, snmpQuerier gosnmpquerier.SyncQuerier) CheckFunction {
	return func() (bool, error, float32) {
		max, err := getMaxValueFromSnmpWalk("1.3.6.1.4.1.2636.3.1.13.1.8", ip, community, snmpQuerier)
		if err == nil {
			return max < maxAllowedTemp, nil, float32(max)
		} else {
			return false, err, 0
		}
	}
}

func NewRabbitMQQueueLenCheck(amqpuri, queue string, max int) CheckFunction {
	return func() (bool, error, float32) {
		queueInfo, err := simpleamqp.NewAmqpManagement(amqpuri).QueueInfo(queue)
		if err == nil {
			return queueInfo.Messages < max, nil, float32(queueInfo.Messages)
		} else {
			return false, nil, float32(0)
		}
	}
}

func NewMysqlConnectionCheck(mysqluri string) CheckFunction {
	return func() (bool, error, float32) {
		u, err := url.Parse(mysqluri)
		if err != nil {
			return false, err, float32(0)
		}

		password, _ := u.User.Password()
		hostAndPort := u.Host
		if !strings.Contains(hostAndPort, ":") {
			hostAndPort = hostAndPort + ":3306"
		}
		con, err := sql.Open("mysql", u.User.Username()+":"+password+"@"+"tcp("+hostAndPort+")"+u.Path)
		defer con.Close()
		if err != nil {
			return false, err, float32(0)
		}
		q := `select CURTIME()`
		row := con.QueryRow(q)
		var date string
		err = row.Scan(&date)
		if err != nil {
			return false, err, float32(0)
		}
		return true, nil, float32(0)
	}
}
