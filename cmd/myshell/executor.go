package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"sync"
)

var builtins = []string{"type", "exit", "echo", "pwd", "cd"}

func execute(commands []Command) {
	if len(commands) == 1 {
		cmd := commands[0]
		defer closeCommandStreams(cmd)

		// Handle exit specially - only exit when it's a standalone command
		if cmd.Name == "exit" {
			code := 0
			if len(cmd.Args) > 0 {
				var err error
				if code, err = strconv.Atoi(cmd.Args[0]); err != nil {
					code = 1
				}
			}
			os.Exit(code)
		}

		executeCmd(cmd)
		return
	}

	executePipeline(commands)
}

func executePipeline(commands []Command) {
	var wg sync.WaitGroup

	for i, command := range commands {
		wg.Add(1)
		go func(idx int, c Command) {
			defer wg.Done()
			defer closeCommandStreams(c)
			executeCmd(c)
		}(i, command)
	}
	wg.Wait()
}

func executeCmd(cmd Command) {
	switch cmd.Name {
	case "exit":
		// In a pipeline, exit is a no-op (it's handled in execute() for standalone commands)
		return

	case "echo":
		fmt.Fprintln(cmd.Stdout, strings.Join(cmd.Args, " "))

	case "type":
		if len(cmd.Args) == 0 {
			return
		}
		if slices.Contains(builtins, cmd.Args[0]) {
			fmt.Fprintln(cmd.Stdout, cmd.Args[0]+" is a shell builtin")
		} else if desc := isExecutableInPath(cmd.Args[0]); desc != "" {
			fmt.Fprintln(cmd.Stdout, desc)
		} else {
			fmt.Fprintln(cmd.Stderr, cmd.Args[0]+": not found")
		}

	case "pwd":
		dir, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(cmd.Stderr, "pwd: error getting current directory")
			return
		}
		fmt.Fprintln(cmd.Stdout, dir)

	case "cd":
		if len(cmd.Args) == 0 {
			return
		}
		p := cmd.Args[0]
		if p == "~" {
			p, _ = os.UserHomeDir()
		}
		if err := os.Chdir(p); err != nil {
			fmt.Fprintf(cmd.Stderr, "cd: %s: No such file or directory\n", cmd.Args[0])
		}

	default:
		c := exec.Command(cmd.Name, cmd.Args...)
		c.Stdin = cmd.Stdin
		c.Stdout = cmd.Stdout
		c.Stderr = cmd.Stderr

		if err := c.Run(); err != nil && isNotFound(err) {
			fmt.Fprintln(cmd.Stderr, cmd.Name+": command not found")
		}
	}
}

func isExecutableInPath(exe string) string {
	for _, dir := range strings.Split(os.Getenv("PATH"), ":") {
		full := dir + "/" + exe
		if info, err := os.Stat(full); err == nil && info.Mode()&0111 != 0 {
			return exe + " is " + full
		}
	}
	return ""
}

func isNotFound(err error) bool {
	return errors.Is(err, exec.ErrNotFound)
}

func closeCommandStreams(cmd Command) {
	closeIfNeeded(cmd.Stdout)
	closeIfNeeded(cmd.Stderr)
	closeIfNeeded(cmd.Stdin)
}

func closeIfNeeded(v any) {
	switch s := v.(type) {
	case *os.File:
		if s != os.Stdout && s != os.Stderr && s != os.Stdin {
			_ = s.Close()
		}
	case io.Closer:
		_ = s.Close()
	}
}
