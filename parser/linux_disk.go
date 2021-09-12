package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

func (parser *LinuxParser) ParseDiskStat(bytesData []byte) (DiskStat, error) {
	var (
		diskInfo = map[string][]string{}
		scanner  = bufio.NewScanner(bytes.NewBuffer(bytesData))
	)

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid line: %s", parts)
		}
		dev := parts[2]
		diskInfo[dev] = parts[3:]
	}

	return diskInfo, scanner.Err()
}
