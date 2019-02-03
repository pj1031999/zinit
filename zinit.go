package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const (
	PREFIX   = "/ram"
	ROOT     = PREFIX + "/root"
	WORKDIR  = PREFIX + "/work"
	UPPERDIR = PREFIX + "/upperdir"
	ROROOT   = "/"
	OLDROOT  = ROOT + "/oldroot"
	PROCFS   = ROOT + "/proc"
	SYSFS    = ROOT + "/sys"
	ZRAMSIZE = "2048M"
)

func main() {
	remountRoot(true)
	createZSWAP()
	createDirs()
	remountRoot(false)
	mountOverlay()
	pivotRoot()
	execInit()
}

func debug(s string) {
	s = "debug>> " + s + "\n"
	_, _ = syscall.Write(1, []byte(s))
}

func createZSWAP() {
	cmd := exec.Command("/sbin/modprobe", "zram")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	cmd = exec.Command("/sbin/zramctl", "--find", "--size", ZRAMSIZE)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	cmd = exec.Command("/sbin/mkfs.ext4", "/dev/zram0")
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func createDir(path string) {
	if _, statErr := os.Stat(path); statErr != nil {
		if err := syscall.Mkdir(path, 0755); err != nil {
			panic(err)
		}
	}
}

func createDirs() {
	createDir(PREFIX)
	if err := syscall.Mount("/dev/zram0", PREFIX, "ext4", uintptr(0), ""); err != nil {
		panic(err)
	}
	createDir(ROOT)
	createDir(WORKDIR)
	createDir(UPPERDIR)
}

func mountOverlay() {
	mOpts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", ROROOT, UPPERDIR, WORKDIR)
	if err := syscall.Mount("tmp", ROOT, "overlay", uintptr(0), mOpts); err != nil {
		panic(err)
	}
	mOpts = fmt.Sprintf("rw,nosuid,nodev,noexec,relatime")
	if err := syscall.Mount("proc", PROCFS, "proc", uintptr(0), mOpts); err != nil {
		panic(err)
	}
	if err := syscall.Mount("sysfs", SYSFS, "sysfs", uintptr(0), mOpts); err != nil {
		panic(err)
	}
}

func pivotRoot() {
	if err := syscall.Mkdir(OLDROOT, 0755); err != nil {
		panic(err)
	}
	if err := syscall.PivotRoot(ROOT, OLDROOT); err != nil {
		panic(err)
	}
}

func remountRoot(rw bool) {
	var flags uintptr
	if rw {
		flags = syscall.MS_REMOUNT
	} else {
		flags = syscall.MS_RDONLY | syscall.MS_REMOUNT
	}
	if err := syscall.Mount("/dev/sda1", "/", "", flags, ""); err != nil {
		panic(err)
	}
}

func execInit() {
	binary, err := exec.LookPath("/bin/sh")
	if err != nil {
		panic(err)
	}
	_, _ = syscall.Write(1, []byte("#########################\n# enter exec /sbin/init #\n#########################\n"))
	if err = syscall.Exec(binary, nil, nil); err != nil {
		panic(err)
	}
}
