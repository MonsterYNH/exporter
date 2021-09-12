package parser

import (
	"github.com/prometheus/procfs/sysfs"
)

func (parser *LinuxParser) ParseNetClass() (sysfs.NetClass, error) {
	return parser.fs.NetClass()
}
