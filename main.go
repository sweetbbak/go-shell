package main

import (
	"bufio"
	"bytes"
	"fmt"
	"golang.org/x/term"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type Termios syscall.Termios

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

func ClearScreen(w io.Writer) (int, error) {
	return w.Write([]byte("\033[H"))
}

func SuspendMe() {
	p, _ := os.FindProcess(os.Getppid())
	p.Signal(syscall.SIGTSTP)
	p, _ = os.FindProcess(os.Getpid())
	p.Signal(syscall.SIGTSTP)
}

func get_stdin() int {
	return syscall.Stdin
}

//	func init_env() {
//		user := os.Getenv("USER")
//		home := os.Getenv("HOME")
//		term := os.Getenv("TERM")
//	}
//
// Move cursor to given position
func moveCursor(x int, y int) {
	var screen *bytes.Buffer = new(bytes.Buffer)
	fmt.Fprintf(screen, "\033[%d;%dH", x, y)
}

// Clear the terminal
func clearTerminal() {
	var output *bufio.Writer = bufio.NewWriter(os.Stdout)
	output.WriteString("\033[2J")
}

func xterminal() error {
	if !term.IsTerminal(0) || !term.IsTerminal(1) {
		return fmt.Errorf("Stdin/Stdout should be a terminal")
	}
	oldState, err := term.MakeRaw(0)
	if err != nil {
		return err
	}
	defer term.Restore(0, oldState)
	screen := struct {
		io.Reader
		io.Writer
	}{os.Stdin, os.Stdout}

	term := term.NewTerminal(screen, "")

	home := os.Getenv("HOME")
	cwd, _ := os.Getwd()
	dir := strings.Replace(cwd, home, "~", 1)
	term.SetPrompt(string(term.Escape.Red) + dir + " $ " + string(term.Escape.Reset))

	// rePrefix := string(term.Escape.Cyan) + "Human says:" + string(term.Escape.Reset)

	for {
		line, err := term.ReadLine()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if line == "" {
			continue
		}
		err = runCommand(line)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		// fmt.Fprintln(term, rePrefix, line)
	}
}

func main() {
	// if err := xterminal(); err != nil {
	// 	fmt.Println("Error: ", err)
	// }
	// var screen *bytes.Buffer = new(bytes.Buffer)
	// var output *bufio.Writer = bufio.NewWriter(os.Stdout)
	// output.WriteString("\033[2J")
	// fmt.Fprintf(screen, "%s", "")

	home := os.Getenv("HOME")
	reader := bufio.NewReader(os.Stdin)
	var buf []byte

	for {
		cwd, _ := os.Getwd()
		dir := strings.Replace(cwd, home, "~", 1)
		fmt.Printf("\033[31m[%s] \033[0m ", dir)

		cmd, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		// byte, err := reader.ReadByte()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		// buf = append(buf, byte)

		if cmd != "\n" {
			err = runCommand(cmd)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
		fmt.Println(string(buf))
	}
}

func runCommand(command string) error {
	command = strings.TrimSuffix(command, "\n")
	args := strings.Fields(command)

	for x := 0; x < len(args); x++ {
		if args[x] == "~" {
			args[x] = os.Getenv("HOME")
		}
		if strings.Contains(args[x], "~") {
			args[x] = strings.ReplaceAll(args[x], "~", os.Getenv("HOME"))
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
		ll := []string{"ls", "-lah", "--color=always"}
		args = append(ll, args[1:]...)

	case "ls":
		ll := []string{"eza", "--icons", "--color=always"}
		args = append(ll, args[1:]...)

	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
