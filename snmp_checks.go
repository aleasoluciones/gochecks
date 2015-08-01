package gochecks

import (
	"time"
)

const (
	sysName = "1.3.6.1.2.1.1.5.0"
)

// SnmpCheckConf snmp connection parameters to use for the check
type SnmpCheckerConf struct {
	retries    int
	timeout    time.Duration
	oidToCheck string
}

// DefaultSnmpCheckConf default values for snmp conection parameters
var DefaultSnmpCheckConf = SnmpCheckerConf{
	retries:    1,
	timeout:    1 * time.Second,
	oidToCheck: sysName,
}

// NewSnmpChecker returns a check function that check if a host respond to a snmp get query
func NewSnmpChecker(host, service, ip, community string, conf SnmpCheckerConf) CheckFunction {
	return func() Event {

		_, err := snmpGet(ip, community, []string{conf.oidToCheck}, conf.timeout, conf.retries)
		if err == nil {
			return Event{Host: host, Service: service, State: "ok", Description: err.Error()}
		}
		return Event{Host: host, Service: service, State: "critical", Description: err.Error()}
	}
}

// NewC4CMTSTempChecker returns a check function that check if any of the slot of a Arris C4 CMTS have a temperature above a given max
func NewC4CMTSTempChecker(host, service, ip, community string, maxAllowedTemp int) CheckFunction {
	return func() Event {

		result, err := snmpWalk(ip, community, "1.3.6.1.4.1.4998.1.1.10.1.4.2.1.29", 2*time.Second, 1)

		if err == nil {
			max := 0
			for _, r := range result {
				if r.Value.(int) != 999 && r.Value.(int) > max {
					max = r.Value.(int)
				}
			}
			var state = "critical"
			if max < maxAllowedTemp {
				state = "ok"
			}
			return Event{Host: host, Service: service, State: state, Metric: float32(max)}
		}
		return Event{Host: host, Service: service, State: "critical", Description: err.Error()}
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
	}
	return 0, err
}

// NewJuniperTempChecker returns a check function that check if a Juniper device (router, switch, etc) have a temperature above a given max
func NewJuniperTempChecker(host, service, ip, community string, maxAllowedTemp uint) CheckFunction {
	return func() Event {
		max, err := getMaxValueFromSnmpWalk("1.3.6.1.4.1.2636.3.1.13.1.7", ip, community)
		if err == nil {
			var state = "critical"
			if max < maxAllowedTemp {
				state = "ok"
			}
			return Event{Host: host, Service: service, State: state, Metric: float32(max)}
		}
		return Event{Host: host, Service: service, State: "critical", Description: err.Error()}
	}
}

// NewJuniperCPUChecker returns a check function that check if a Juniper device (router, switch, etc) have a any cpu usage above a given percent
func NewJuniperCPUChecker(host, service, ip, community string, maxAllowedCPUPercent uint) CheckFunction {
	return func() Event {
		max, err := getMaxValueFromSnmpWalk("1.3.6.1.4.1.2636.3.1.13.1.8", ip, community)
		if err == nil {
			var state = "critical"
			if max < maxAllowedCPUPercent {
				state = "ok"
			}
			return Event{Host: host, Service: service, State: state, Metric: float32(max)}
		}
		return Event{Host: host, Service: service, State: "critical", Description: err.Error()}
	}
}
