# containerception

---

## What is a container?

A container is a standard unit of software that packages up code and all its dependencies so the application runs quickly and reliably from one computing environment to another. In an ideal situation situation the container should also limit what the process can do to the host machine. This is usually accomplished by control groups and namespaces. Namespaces restrict the containers access to resources such as the network and hostnames. Control groups restrict access physical resources such as CPU, memory, and disk.
[Docker](https://www.docker.com/resources/what-container). [Namespaces & CGroups](https://medium.com/@teddyking/linux-namespaces-850489d3ccf)

## Fork & Exec

In order to create the container, we need to run a new command and use fork/exec to setup a child process with the correct permissions. The child process that is run will technically be the container, with the parent process being solely responsible for setting up the correct environment. It's a bit of a hack, but is also used by the talented folks at docker.

## First pass without any namespaces or control groups.

Try running the below golang application using `go run main.go run echo "FOO"`

```
package main

import (
	"fmt"
	"os"
	"os/exec"
)

//need to run at least the following
// cp /bin/ -R /home/$USER/rootfs/bin
// mkdir /home/$USER/rootfs/proc
func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("Invalid command")
	}
}
func run() {
	//fork and exec
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run())
}
func child() {
	fmt.Printf("Running %v with PID %v\n", os.Args[2:], os.Getpid())
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(cmd.Run())
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}

```

Now try `go run main.go run echo "HELLO" > test.txt` notice after we run the command the file is there?
Now try `go run main.go run ps aux` notice that the container can see all the hosts running processes?

Let's lock this down by setting up some namespaces

## Golang implemention

```
package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

//need to run at least the following
// cp /bin/ -R /home/$USER/rootfs/bin
// mkdir /home/$USER/rootfs/proc
//
func main() {
	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		panic("Invalid command")
	}
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
	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//setup hostname etc
	must(syscall.Sethostname([]byte("containerception")))
	must(cmd.Run())
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}

```

Notice the PID is 1. That's because the syscall flags setup a new PID, Network, and UTS.
Now try `go run main.go run ps aux` notice that the container can _still_ see all the hosts running processes? That's because ps is special and reads in the processes in the /proc filesystem. So we need to make that.
Try `go run main.go run /bin/bash` notice the terminal prompt is now `root@containerception` that's because we have setup a new hostname. Please feel free to play around in the shell for a bit, and make sure to notice how the changes you made are evident after you exit out (Ctrl+D).

Let's setup the proc folder for ps as well as ensure the container is restricted to it's own file system. PS. time to bust out CGroups.

```
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
//hacky way to copy directories for now.. This is solely for ubuntu and definitely not recommended for production
func initialize() {
        must(MakeDir("/rootfs"))
        must(MakeDir("/rootfs/proc"))
        must(CopyDir("/bin/", "/rootfs/bin/"))
        must(CopyDir("/dev/", "/rootfs/dev/"))
        must(CopyDir("/etc/", "/rootfs/etc/"))
        must(CopyDir("/lib/", "/rootfs/lib/"))
        must(CopyDir("/root/", "/rootfs/root/"))
        must(CopyDir("/tmp/", "/rootfs/tmp/"))
        must(CopyDir("/usr/", "/rootfs/usr/"))
        must(CopyDir("/usr/", "/rootfs/usr/"))
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
        must(syscall.Unmount("proc",0))
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


```
