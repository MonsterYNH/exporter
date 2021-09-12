package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var (
	procNetDevInterfaceRE = regexp.MustCompile(`^(.+): *(.+)$`)
	procNetDevFieldSep    = regexp.MustCompile(` +`)
	// newNetDevFilter       = newNetDevFilter("", "")
)

// type netDevFilter struct {
// 	ignorePattern *regexp.Regexp
// 	acceptPattern *regexp.Regexp
// }

// func newNetDevFilter(ignoredPattern, acceptPattern string) (f netDevFilter) {
// 	if ignoredPattern != "" {
// 		f.ignorePattern = regexp.MustCompile(ignoredPattern)
// 	}

// 	if acceptPattern != "" {
// 		f.acceptPattern = regexp.MustCompile(acceptPattern)
// 	}

// 	return
// }

func (parser *LinuxParser) ParseNetStat(bytesData []byte) (NetStat, error) {
	scanner := bufio.NewScanner(bytes.NewBuffer(bytesData))
	scanner.Scan() // skip first header
	scanner.Scan()
	parts := strings.Split(scanner.Text(), "|")
	if len(parts) != 3 { // interface + receive + transmit
		return nil, fmt.Errorf("invalid header line in net/dev: %s",
			scanner.Text())
	}

	receiveHeader := strings.Fields(parts[1])
	transmitHeader := strings.Fields(parts[2])
	headerLength := len(receiveHeader) + len(transmitHeader)

	netDev := NetStat{}
	for scanner.Scan() {
		line := strings.TrimLeft(scanner.Text(), " ")
		parts := procNetDevInterfaceRE.FindStringSubmatch(line)
		if len(parts) != 3 {
			return nil, fmt.Errorf("couldn't get interface name, invalid line in net/dev: %q", line)
		}

		dev := parts[1]
		// if filter.ignored(dev) {
		// 	level.Debug(logger).Log("msg", "Ignoring device", "device", dev)
		// 	continue
		// }

		values := procNetDevFieldSep.Split(strings.TrimLeft(parts[2], " "), -1)
		if len(values) != headerLength {
			return nil, fmt.Errorf("couldn't get values, invalid line in net/dev: %q", parts[2])
		}

		devStats := map[string]uint64{}
		addStats := func(key, value string) {
			v, err := strconv.ParseUint(value, 0, 64)
			if err != nil {
				log.Println("invalid value in netstats, key:", key, ", value:", value, ", error:", err)
				return
			}

			devStats[key] = v
		}

		for i := 0; i < len(receiveHeader); i++ {
			addStats("receive_"+receiveHeader[i], values[i])
		}

		for i := 0; i < len(transmitHeader); i++ {
			addStats("transmit_"+transmitHeader[i], values[i+len(receiveHeader)])
		}

		netDev[dev] = devStats
	}
	return netDev, scanner.Err()
}
