package parser

import (
	"bufio"
	"bytes"
	"context"
	"exporter/util"
	"fmt"
	"log"
	"strings"
)

func (parser *LinuxParser) ParseNetStatInfo() (map[string]map[string]string, error) {
	bytes, err := util.ExecCommand(context.Background(), "cat", "/proc/net/netstat")
	if err != nil {
		return nil, err
	}
	netStats, err := parseNetStats(bytes)
	if err != nil {
		return nil, err
	}
	bytes, err = util.ExecCommand(context.Background(), "cat", "/proc/net/snmp")
	if err != nil {
		return nil, err
	}
	snmpStats, err := parseNetStats(bytes)
	if err != nil {
		return nil, err
	}

	bytes, err = util.ExecCommand(context.Background(), "cat", "/proc/net/snmp6")
	if err != nil {
		log.Println(err)
	} else {
		snmp6Stats, err := parseSNMP6Stats(bytes)
		if err != nil {
			return nil, err
		}
		for k, v := range snmp6Stats {
			netStats[k] = v
		}
	}

	// Merge the results of snmpStats into netStats (collisions are possible, but
	// we know that the keys are always unique for the given use case).
	for k, v := range snmpStats {
		netStats[k] = v
	}

	return netStats, nil
}

func parseNetStats(bytesData []byte) (map[string]map[string]string, error) {
	var (
		netStats = map[string]map[string]string{}
		scanner  = bufio.NewScanner(bytes.NewBuffer(bytesData))
	)

	for scanner.Scan() {
		nameParts := strings.Split(scanner.Text(), " ")
		scanner.Scan()
		valueParts := strings.Split(scanner.Text(), " ")
		// Remove trailing :.
		protocol := nameParts[0][:len(nameParts[0])-1]
		netStats[protocol] = map[string]string{}
		if len(nameParts) != len(valueParts) {
			return nil, fmt.Errorf("mismatch field count mismatch : %s", protocol)
		}
		for i := 1; i < len(nameParts); i++ {
			netStats[protocol][nameParts[i]] = valueParts[i]
		}
	}

	return netStats, scanner.Err()
}

func parseSNMP6Stats(bytesData []byte) (map[string]map[string]string, error) {
	var (
		netStats = map[string]map[string]string{}
		scanner  = bufio.NewScanner(bytes.NewBuffer(bytesData))
	)

	for scanner.Scan() {
		stat := strings.Fields(scanner.Text())
		if len(stat) < 2 {
			continue
		}
		// Expect to have "6" in metric name, skip line otherwise
		if sixIndex := strings.Index(stat[0], "6"); sixIndex != -1 {
			protocol := stat[0][:sixIndex+1]
			name := stat[0][sixIndex+1:]
			if _, present := netStats[protocol]; !present {
				netStats[protocol] = map[string]string{}
			}
			netStats[protocol][name] = stat[1]
		}
	}

	return netStats, scanner.Err()
}
