package systems

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const HasNTService = false
const HasLaunchd = false
const HasLinuxBackends = true

func init() {

}

func Interactive() bool {
	yes, err := isInteractive()
	return yes && err == nil
}

func InContainer() bool {
	yes, err := isInContainer(cgroupFile)
	return yes && err == nil
}

func isInteractive() (yes bool, err error) {
	yes, err = isInContainer(cgroupFile)
	if yes || err != nil {
		return
	}

	ppid := os.Getppid()
	if ppid == 1 {
		return
	}

	binary, _ := binaryName(ppid)
	yes = binary != "systemd"
	return
}

// isInContainer checks if the service is being executed in
// docker or lxc container.
func isInContainer(cgroupPath string) (yes bool, err error) {
	var f *os.File
	f, err = os.Open(cgroupPath)
	if err != nil {
		return
	}
	defer f.Close()

	const linesThreshold = 8
	lines := 0
	scan := bufio.NewScanner(f)
	for scan.Scan() && lines <= linesThreshold {
		if yes = strings.Contains(scan.Text(), "docker") ||
			strings.Contains(scan.Text(), "lxc"); yes {
			return
		}
		lines++
	}

	err = scan.Err()
	return
}

func binaryName(pid int) (name string, err error) {
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	var data []byte
	data, err = os.ReadFile(statPath)
	if err != nil {
		return
	}

	// First, parse out the image name
	content := string(data)
	binStart := strings.IndexRune(content, '(') + 1
	binEnd := strings.IndexRune(content[binStart:], ')')
	name = content[binStart : binStart+binEnd]
	return
}

const cgroupFile = "/proc/1/cgroup"
