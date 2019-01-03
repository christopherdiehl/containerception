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
