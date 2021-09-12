package parser

import (
	"bytes"
	"context"
	"exporter/util"
	"fmt"
)

func (parser *LinuxParser) ParseFileFDStat() (map[string]string, error) {
	bytesData, err := util.ExecCommand(context.Background(), "cat", "/proc/sys/fs/file-nr")
	if err != nil {
		return nil, err
	}

	parts := bytes.Split(bytes.TrimSpace(bytesData), []byte("\u0009"))
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected number of file stats in %q", "/proc/sys/fs/file-nr")
	}

	var fileFDStat = map[string]string{}
	// The file-nr proc is only 1 line with 3 values.
	fileFDStat["allocated"] = string(parts[0])
	// The second value is skipped as it will always be zero in linux 2.6.
	fileFDStat["maximum"] = string(parts[2])

	return fileFDStat, nil
}
