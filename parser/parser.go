package parser

import "github.com/prometheus/procfs/sysfs"

type Parser interface {
	ParseCPUStat([]byte) (Stat, error)
	ParseDiskStat([]byte) (DiskStat, error)
	ParseFileSystemStat([]byte) ([]FileSystemStat, error)
	ParseIPStat() (IPStat, error)
	ParseLoadAvgStat([]byte) (LoadAvgStat, error)
	ParseMemoryStat([]byte) (MemoryStat, error)
	ParseNetStat([]byte) (NetStat, error)
	ParseUname() (UnameStat, error)
	ParseBootTime() (float64, error)
	ParseNetClass() (sysfs.NetClass, error)
	ParseNetStatInfo() (map[string]map[string]string, error)
	ParseFileFDStat() (map[string]string, error)
}

var sysPath = "/sys"

type LinuxParser struct {
	fs sysfs.FS
}

func NewLinuxParser() (*LinuxParser, error) {
	fs, err := sysfs.NewFS(sysPath)
	if err != nil {
		return nil, err
	}

	return &LinuxParser{fs: fs}, nil
}

// CPUStat shows how much time the cpu spend in various stages.
type CPUStat struct {
	User      float64
	Nice      float64
	System    float64
	Idle      float64
	Iowait    float64
	IRQ       float64
	SoftIRQ   float64
	Steal     float64
	Guest     float64
	GuestNice float64
}

// SoftIRQStat represent the softirq statistics as exported in the procfs stat file.
// A nice introduction can be found at https://0xax.gitbooks.io/linux-insides/content/interrupts/interrupts-9.html
// It is possible to get per-cpu stats by reading /proc/softirqs
type SoftIRQStat struct {
	Hi          uint64
	Timer       uint64
	NetTx       uint64
	NetRx       uint64
	Block       uint64
	BlockIoPoll uint64
	Tasklet     uint64
	Sched       uint64
	Hrtimer     uint64
	Rcu         uint64
}

// Stat represents kernel/system statistics.
type Stat struct {
	// Boot time in seconds since the Epoch.
	BootTime uint64
	// Summed up cpu statistics.
	CPUTotal CPUStat
	// Per-CPU statistics.
	CPU []CPUStat
	// Number of times interrupts were handled, which contains numbered and unnumbered IRQs.
	IRQTotal uint64
	// Number of times a numbered IRQ was triggered.
	IRQ []uint64
	// Number of times a context switch happened.
	ContextSwitches uint64
	// Number of times a process was created.
	ProcessCreated uint64
	// Number of processes currently running.
	ProcessesRunning uint64
	// Number of processes currently blocked (waiting for IO).
	ProcessesBlocked uint64
	// Number of times a softirq was scheduled.
	SoftIRQTotal uint64
	// Detailed softirq statistics.
	SoftIRQ SoftIRQStat
}

type DiskStat map[string][]string

type FileSystemStat struct {
	Labels            FileSystemLabels
	Size, Free, Avail float64
	Files, FilesFree  float64
	Ro, DeviceError   float64
}

type FileSystemLabels struct {
	Device, MountPoint, FsType, Options string
}

type IPStat []string

type LoadAvgStat []float64

type MemoryStat map[string]float64

type NetStat map[string]map[string]uint64

type UnameStat struct {
	SysName    string
	Release    string
	Version    string
	Machine    string
	NodeName   string
	DomainName string
}
