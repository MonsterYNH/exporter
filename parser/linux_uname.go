package parser

import (
	"bytes"

	"golang.org/x/sys/unix"
)

func (parser *LinuxParser) ParseUname() (UnameStat, error) {
	var utsname unix.Utsname
	if err := unix.Uname(&utsname); err != nil {
		return UnameStat{}, err
	}

	output := UnameStat{
		SysName:    string(utsname.Sysname[:bytes.IndexByte(utsname.Sysname[:], 0)]),
		Release:    string(utsname.Release[:bytes.IndexByte(utsname.Release[:], 0)]),
		Version:    string(utsname.Version[:bytes.IndexByte(utsname.Version[:], 0)]),
		Machine:    string(utsname.Machine[:bytes.IndexByte(utsname.Machine[:], 0)]),
		NodeName:   string(utsname.Nodename[:bytes.IndexByte(utsname.Nodename[:], 0)]),
		DomainName: string(utsname.Domainname[:bytes.IndexByte(utsname.Domainname[:], 0)]),
	}

	return output, nil
}
