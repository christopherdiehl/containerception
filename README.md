# containerception

---

## What is a container?

A container is a standard unit of software that packages up code and all its dependencies so the application runs quickly and reliably from one computing environment to another. In an ideal situation situation the container should also limit what the process can do to the host machine. This is usually accomplished by control groups and namespaces. Namespaces restrict the containers access to resources such as the network and hostnames. Control groups restrict access physical resources such as CPU, memory, and disk.
[Docker](https://www.docker.com/resources/what-container). [Namespaces & CGroups](https://medium.com/@teddyking/linux-namespaces-850489d3ccf)

## Fork & Exec

In order to create the container, we need to run a new command and use fork/exec to setup a child process with the correct permissions. The child process that is run will technically be the container, with the parent process being solely responsible for setting up the correct environment. It's a bit of a hack, but we'll make it work.

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
		panic("Invalid command")x
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

Let's setup the proc folder for ps as well as ensure the container is restricted to it's own file system.

```
package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

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
func must(err error) {
	if err != nil {
		panic(err)
	}
}


```

Play around with that and you should see that ps is working correctly and the container has it's own little filesystem! Please note this filesystem persists across multiple containers and be aware that the rootfs directory is created in the Dockerfile. If you're interested I would recommend you try to expirement with the capaibilites of the container using `go run main.go run /bin/bash` then trying something like a fork bomb: `bomb(){ bomb|bomb& };bomb` (Note be careful with that thing)

Credit for this repo goes to the amazing namespaces in Go articles by [Teddy King](https://github.com/teddyking) as well as Liz Rice's own talk [Container From Scratch](https://github.com/lizrice/containers-from-scratch)
