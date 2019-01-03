# containerception

---

## What is a container?

A container is a standard unit of software that packages up code and all its dependencies so the application runs quickly and reliably from one computing environment to another. In an ideal situation situation the container should also limit what the process can do to the host machine. This is usually accomplished by control groups and namespaces. Namespaces restrict the containers access to resources such as the network and hostnames. Control groups restrict access physical resources such as CPU, memory, and disk.
[Docker](https://www.docker.com/resources/what-container). [Namespaces & CGroups](https://medium.com/@teddyking/linux-namespaces-850489d3ccf)

## Fork & Exec

In order to create the container, we need to run a new command and use fork/exec to setup a child process with the correct permissions. The child process that is run will technically be the container, with the parent process being solely responsible for setting up the correct environment. It's a bit of a hack, but is also used by the talented folks at docker.

## First pass without any namespaces or control groups.

Try running the below golang application using `go run main.go run echo "FOO"`
Now try `go run main.go run

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
	must(syscall.Sethostname([]byte("cdock")))
	must(cmd.Run())
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
func OSReadDir(root string) ([]string, error) {

	var files []string
	f, err := os.Open(root)
	if err != nil {
		return files, err
	}
	fileInfo, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return files, err
	}

	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return files, nil
}
```
