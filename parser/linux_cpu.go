package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	userHZ = 100.0
)

func (parser *LinuxParser) ParseCPUStat(bytesData []byte) (Stat, error) {
	stat := Stat{
		CPU: make([]CPUStat, 0),
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(bytesData))

	var err error
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		if len(parts) < 2 {
			continue
		}

		name := strings.ToLower(parts[0])
		switch {
		case strings.HasPrefix(name, "cpu"):
			id, cpuStat, err := parseCPUStat(line)
			if err != nil {
				return stat, err
			}
			if id == -1 {
				stat.CPUTotal = cpuStat
			} else {
				for len(stat.CPU) <= id {
					stat.CPU = append(stat.CPU, CPUStat{})
				}
				stat.CPU[id] = cpuStat
			}
		case name == "ctxt":
			if stat.ContextSwitches, err = strconv.ParseUint(parts[1], 10, 64); err != nil {
				return stat, fmt.Errorf("couldn't parse %s (ctxt): %s", parts[1], err)
			}
		case name == "btime":
			if stat.BootTime, err = strconv.ParseUint(parts[1], 10, 64); err != nil {
				return stat, fmt.Errorf("couldn't parse %s (btime): %s", parts[1], err)
			}
		case name == "processes":
			if stat.ProcessCreated, err = strconv.ParseUint(parts[1], 10, 64); err != nil {
				return stat, fmt.Errorf("couldn't parse %s (processes): %s", parts[1], err)
			}
		case name == "procs_running":
			if stat.ProcessesRunning, err = strconv.ParseUint(parts[1], 10, 64); err != nil {
				return stat, fmt.Errorf("couldn't parse %s (procs_running): %s", parts[1], err)
			}
		case name == "procs_blocked":
			if stat.ProcessesBlocked, err = strconv.ParseUint(parts[1], 10, 64); err != nil {
				return stat, fmt.Errorf("couldn't parse %s (procs_blocked): %s", parts[1], err)
			}
		case name == "softirq":
			total, softIRQStats, err := parseSoftIRQStat(line)
			if err != nil {
				return stat, err
			}
			stat.SoftIRQTotal = total
			stat.SoftIRQ = softIRQStats
		}
	}

	return stat, scanner.Err()
}

func parseCPUStat(data string) (int, CPUStat, error) {
	var cpuStat CPUStat
	var cpu string
	count, err := fmt.Sscanf(
		data,
		"%s %f %f %f %f %f %f %f %f %f %f",
		&cpu,
		&cpuStat.User, &cpuStat.Nice, &cpuStat.System, &cpuStat.Idle,
		&cpuStat.Iowait, &cpuStat.IRQ, &cpuStat.SoftIRQ, &cpuStat.Steal,
		&cpuStat.Guest, &cpuStat.GuestNice,
	)

	cpuStat.User /= userHZ
	cpuStat.Nice /= userHZ
	cpuStat.System /= userHZ
	cpuStat.Idle /= userHZ
	cpuStat.Iowait /= userHZ
	cpuStat.IRQ /= userHZ
	cpuStat.SoftIRQ /= userHZ
	cpuStat.Steal /= userHZ
	cpuStat.Guest /= userHZ
	cpuStat.GuestNice /= userHZ

	if err != nil && err != io.EOF {
		return -1, cpuStat, fmt.Errorf("couldn't parse %s (cpu): %s", data, err)
	}
	if count == 0 {
		return -1, cpuStat, fmt.Errorf("couldn't parse %s (cpu): 0 elements parsed", data)
	}

	if cpu == "cpu" {
		return -1, cpuStat, nil
	}

	cpuID, err := strconv.Atoi(cpu[3:])
	if err != nil {
		return -1, cpuStat, fmt.Errorf("couldn't parse %s (cpu/cpuid): %s", data, err)
	}

	return cpuID, cpuStat, nil
}

// Parse a softirq line.
func parseSoftIRQStat(line string) (uint64, SoftIRQStat, error) {
	softIRQStat := SoftIRQStat{}
	var total uint64
	var prefix string

	_, err := fmt.Sscanf(line, "%s %d %d %d %d %d %d %d %d %d %d %d",
		&prefix, &total,
		&softIRQStat.Hi, &softIRQStat.Timer, &softIRQStat.NetTx, &softIRQStat.NetRx,
		&softIRQStat.Block, &softIRQStat.BlockIoPoll,
		&softIRQStat.Tasklet, &softIRQStat.Sched,
		&softIRQStat.Hrtimer, &softIRQStat.Rcu)

	if err != nil {
		return total, softIRQStat, fmt.Errorf("couldn't parse %s (softirq): %s", line, err)
	}

	return total, softIRQStat, nil
}
