package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func System(cmd string) int {
	c := exec.Command("sh", "-c", cmd)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()

	if err == nil {
		return 0
	}

	// Figure out the exit code
	if ws, ok := c.ProcessState.Sys().(syscall.WaitStatus); ok {
		if ws.Exited() {
			return ws.ExitStatus()
		}

		if ws.Signaled() {
			return -int(ws.Signal())
		}
	}

	return -1
}

func sys_exec() {
	binary, lookErr := exec.LookPath("ls")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"ls", "-a", "-l", "-h", "--color=always", "/home/sweet"}

	env := os.Environ()

	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		fmt.Println(execErr)
		// panic(execErr)
	}
	os.Exit(0)
}

func main() {

	reader := bufio.NewReader(os.Stdin)
	for {
		dir, _ := os.Getwd()
		fmt.Printf("[%s] $ ", dir)
		cmdString, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		err = runCommand(cmdString)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func runCommand(command string) error {
	command = strings.TrimSuffix(command, "\n")
	args := strings.Fields(command)

	for x := 0; x < len(args); x++ {
		if args[x] == "~" {
			args[x] = os.Getenv("HOME")
		}
	}

	switch args[0] {
	case "exit":
		os.Exit(0)

	case "cd":
		if len(args) < 2 {
			return os.Chdir(os.Getenv("HOME"))
		}
		return os.Chdir(args[1])

	case "ll":
		ll := []string{"ls", "-l", "--color=always"}
		args = append(ll, args[1:]...)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
