package parser

import (
	"context"
	"exporter/util"
	"strings"
	"time"
)

func (parser *LinuxParser) ParseBootTime() (float64, error) {
	cmd := `date -d "$(awk -F. '{print $1}' /proc/uptime) second ago" +"%Y-%m-%d %H:%M:%S"`

	bytes, err := util.ExecCommand(context.Background(), "bash", "-c", cmd)
	if err != nil {
		return 0, err
	}
	date, err := time.Parse("2006-01-02 15:04:05", strings.ReplaceAll(string(bytes), "\n", ""))
	if err != nil {
		return 0, err
	}

	return float64(date.Unix()), nil
}
