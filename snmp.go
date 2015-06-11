package gochecks

import (
	"time"

	"github.com/soniah/gosnmp"
)

func snmpWalk(destination, community, oid string, timeout time.Duration, retries int) ([]gosnmp.SnmpPDU, error) {
	conn := snmpConnection(destination, community, timeout, retries)
	if err := conn.Connect(); err != nil {
		return nil, err
	}
	defer conn.Conn.Close()

	return conn.BulkWalkAll(oid)
}

func snmpGet(destination, community string, oids []string, timeout time.Duration, retries int) ([]gosnmp.SnmpPDU, error) {
	conn := snmpConnection(destination, community, timeout, retries)
	if err := conn.Connect(); err != nil {
		return nil, err
	}
	defer conn.Conn.Close()

	result, err := conn.Get(oids)
	if err != nil {
		return nil, err
	}

	pdus := []gosnmp.SnmpPDU{}
	for _, pdu := range result.Variables {
		pdus = append(pdus, pdu)
	}
	return pdus, nil

}

func snmpConnection(destination, community string, timeout time.Duration, retries int) gosnmp.GoSNMP {
	return gosnmp.GoSNMP{
		Target:    destination,
		Port:      161,
		Community: community,
		Version:   gosnmp.Version2c,
		Timeout:   timeout,
		Retries:   retries,
	}
}
