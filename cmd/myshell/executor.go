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

var builtins = []string{"type", "exit", "echo", "pwd", "cd", "history"}

// ExitError is returned when the shell's exit builtin is invoked.
type ExitError struct{ Code int }

func (e *ExitError) Error() string { return fmt.Sprintf("exit: %d", e.Code) }

var builtinHandlers = map[string]func(Command) error{
	"exit":    builtinExit,
	"echo":    builtinEcho,
	"type":    builtinType,
	"pwd":     builtinPwd,
	"cd":      builtinCd,
	"history": builtinHistory,
}

func builtinHistory(cmd Command) error {
	return nil
}

func builtinExit(cmd Command) error {
	code := 0
	if len(cmd.Args) > 0 {
		var err error
		if code, err = strconv.Atoi(cmd.Args[0]); err != nil {
			code = 1
		}
	}
	return &ExitError{code}
}

func builtinEcho(cmd Command) error {
	fmt.Fprintln(cmd.Stdout, strings.Join(cmd.Args, " "))
	return nil
}

func builtinType(cmd Command) error {
	if len(cmd.Args) == 0 {
		return nil
	}
	if slices.Contains(builtins, cmd.Args[0]) {
		fmt.Fprintln(cmd.Stdout, cmd.Args[0]+" is a shell builtin")
	} else if desc := isExecutableInPath(cmd.Args[0]); desc != "" {
		fmt.Fprintln(cmd.Stdout, desc)
	} else {
		fmt.Fprintln(cmd.Stderr, cmd.Args[0]+": not found")
	}
	return nil
}

func builtinPwd(cmd Command) error {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(cmd.Stderr, "pwd: error getting current directory")
		return nil
	}
	fmt.Fprintln(cmd.Stdout, dir)
	return nil
}

func builtinCd(cmd Command) error {
	if len(cmd.Args) == 0 {
		return nil
	}
	p := cmd.Args[0]
	if p == "~" {
		p, _ = os.UserHomeDir()
	}
	if err := os.Chdir(p); err != nil {
		fmt.Fprintf(cmd.Stderr, "cd: %s: No such file or directory\n", cmd.Args[0])
	}
	return nil
}

func execute(commands []Command) error {
	if len(commands) == 1 {
		cmd := commands[0]
		defer closeCommandStreams(cmd)
		return executeCmd(cmd)
	}
	return executePipeline(commands)
}

func executePipeline(commands []Command) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i, command := range commands {
		wg.Add(1)
		go func(idx int, c Command) {
			defer wg.Done()
			defer closeCommandStreams(c)
			if err := executeCmd(c); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}(i, command)
	}
	wg.Wait()
	return firstErr
}

func executeCmd(cmd Command) error {
	if handler, ok := builtinHandlers[cmd.Name]; ok {
		return handler(cmd)
	}

	c := exec.Command(cmd.Name, cmd.Args...)
	c.Stdin = cmd.Stdin
	c.Stdout = cmd.Stdout
	c.Stderr = cmd.Stderr

	if err := c.Run(); err != nil && isNotFound(err) {
		fmt.Fprintln(cmd.Stderr, cmd.Name+": command not found")
	}
	return nil
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
