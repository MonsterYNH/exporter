package parser

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	procPath = "/proc"
)

func (parser *LinuxParser) ParseLoadAvgStat(bytes []byte) (LoadAvgStat, error) {
	loads := make([]float64, 3)
	parts := strings.Fields(string(bytes))
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected content in %s", procFilePath("loadavg"))
	}

	var err error
	for i, load := range parts[0:3] {
		loads[i], err = strconv.ParseFloat(load, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse load '%s': %w", load, err)
		}
	}
	return loads, nil
}

func procFilePath(name string) string {
	return filepath.Join(procPath, name)
}
