package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

const (
	rootfsPath             = "/"
	defMountPointsExcluded = "^/(dev|proc|sys|var/lib/docker/.+)($|/)"
	defFSTypesExcluded     = "^(autofs|binfmt_misc|bpf|cgroup2?|configfs|debugfs|devpts|devtmpfs|fusectl|hugetlbfs|iso9660|mqueue|nsfs|overlay|proc|procfs|pstore|rpc_pipefs|securityfs|selinuxfs|squashfs|sysfs|tracefs)$"
)

var (
	excludedMountPointsPattern = regexp.MustCompile(defMountPointsExcluded)
	excludedFSTypesPattern     = regexp.MustCompile(defFSTypesExcluded)
	mountTimeout               = 5
	stuckMounts                = make(map[string]struct{})
	stuckMountsMtx             = &sync.Mutex{}
)

func (parser *LinuxParser) ParseFileSystemStat(bytes []byte) ([]FileSystemStat, error) {
	mps, err := parseFilesystemLabels(bytes)
	if err != nil {
		return nil, err
	}

	stats := make([]FileSystemStat, 0)

	for _, labels := range mps {
		if excludedMountPointsPattern.MatchString(labels.MountPoint) {
			continue
		}
		if excludedFSTypesPattern.MatchString(labels.FsType) {
			continue
		}

		stuckMountsMtx.Lock()

		if _, ok := stuckMounts[labels.MountPoint]; ok {
			stats = append(stats, FileSystemStat{
				Labels:      labels,
				DeviceError: 1,
			})
			log.Println("[ERRPR] Mount point is in an unresponsive state, mountpoint: " + labels.MountPoint)
			stuckMountsMtx.Unlock()
			continue
		}
		stuckMountsMtx.Unlock()

		success := make(chan struct{})
		go stuckMountWatcher(labels.MountPoint, success)

		buf := new(unix.Statfs_t)
		err = unix.Statfs(rootfsFilePath(labels.MountPoint), buf)
		stuckMountsMtx.Lock()
		close(success)
		// If the mount has been marked as stuck, unmark it and log it's recovery.
		if _, ok := stuckMounts[labels.MountPoint]; ok {
			log.Println("Mount point has recovered, monitoring will resume", "mountpoint", labels.MountPoint)
			delete(stuckMounts, labels.MountPoint)
		}
		stuckMountsMtx.Unlock()

		if err != nil {
			stats = append(stats, FileSystemStat{
				Labels:      labels,
				DeviceError: 1,
			})

			log.Println("Error on statfs() system call, rootfs:" + labels.MountPoint)
			continue
		}

		var ro float64
		for _, option := range strings.Split(labels.Options, ",") {
			if option == "ro" {
				ro = 1
				break
			}
		}

		stats = append(stats, FileSystemStat{
			Labels:    labels,
			Size:      float64(buf.Blocks) * float64(buf.Bsize),
			Free:      float64(buf.Bfree) * float64(buf.Bsize),
			Avail:     float64(buf.Bavail) * float64(buf.Bsize),
			Files:     float64(buf.Files),
			FilesFree: float64(buf.Ffree),
			Ro:        ro,
		})

	}
	return stats, nil
}

// stuckMountWatcher listens on the given success channel and if the channel closes
// then the watcher does nothing. If instead the timeout is reached, the
// mount point that is being watched is marked as stuck.
func stuckMountWatcher(mountPoint string, success chan struct{}) {
	select {
	case <-success:
		// Success
	case <-time.After(time.Duration(mountTimeout)):
		// Timed out, mark mount as stuck
		stuckMountsMtx.Lock()
		select {
		case <-success:
			// Success came in just after the timeout was reached, don't label the mount as stuck
		default:
			log.Println("Mount point timed out, it is being labeled as stuck and will not be monitored, mountpoint:", mountPoint)
			stuckMounts[mountPoint] = struct{}{}
		}
		stuckMountsMtx.Unlock()
	}
}

func parseFilesystemLabels(bytesData []byte) ([]FileSystemLabels, error) {
	var filesystems []FileSystemLabels

	scanner := bufio.NewScanner(bytes.NewBuffer(bytesData))
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())

		if len(parts) < 4 {
			return nil, fmt.Errorf("malformed mount point information: %q", scanner.Text())
		}

		// Ensure we handle the translation of \040 and \011
		// as per fstab(5).
		parts[1] = strings.Replace(parts[1], "\\040", " ", -1)
		parts[1] = strings.Replace(parts[1], "\\011", "\t", -1)

		filesystems = append(filesystems, FileSystemLabels{
			Device:     parts[0],
			MountPoint: rootfsStripPrefix(parts[1]),
			FsType:     parts[2],
			Options:    parts[3],
		})
	}

	return filesystems, scanner.Err()
}

func rootfsStripPrefix(path string) string {
	if rootfsPath == "/" {
		return path
	}
	stripped := strings.TrimPrefix(path, rootfsPath)
	if stripped == "" {
		return "/"
	}
	return stripped
}

func rootfsFilePath(name string) string {
	return filepath.Join(rootfsPath, name)
}
