package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"syscall"
)

//need to run at least the following
// cp /bin/ -R /home/$USER/rootfs/bin
// mkdir /home/$USER/rootfs/proc
//
func main() {
	switch os.Args[1] {
	case "init":
		initialize()
	case "run":
		run()
	case "child":
		child()
	default:
		panic("Invalid command")
	}
}
func initialize() {
	must(MakeDir("/rootfs"))
	must(MakeDir("/rootfs/proc"))
	must(CopyDir("/bin/", "/rootfs/bin/"))
}
func run() {
	//fork and exec
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	must(cmd.Run())
}
func child() {
	fmt.Printf("Running %v with PID %v\n", os.Args[2:], os.Getpid())
	must(syscall.Mount("rootfs", "rootfs", "", syscall.MS_BIND, ""))
	must(syscall.Chroot("rootfs"))
	must(os.Chdir("/"))
	must(syscall.Mount("proc", "proc", "proc", 0, ""))
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//setup hostname etc
	must(syscall.Sethostname([]byte("containerception")))
	must(cmd.Run())
	must(syscall.Unmount("proc", 0))
}
func CopyDir(src, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = CopyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
func CopyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}
func MakeDir(dst string) error {
	cmd := exec.Command("mkdir", dst)
	return cmd.Run()
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
