package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// ExitError is returned when the shell's exit builtin is invoked.
type ExitError struct{ Code int }

func (e *ExitError) Error() string { return fmt.Sprintf("exit: %d", e.Code) }

func (s *Shell) execute(commands []Command) error {
	if len(commands) == 1 {
		cmd := commands[0]
		defer closeCommandStreams(cmd)
		return s.executeCmd(cmd)
	}
	return s.executePipeline(commands)
}

func (s *Shell) executePipeline(commands []Command) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for _, command := range commands {
		wg.Add(1)
		go func(c Command) {
			defer wg.Done()
			defer closeCommandStreams(c)
			if err := s.executeCmd(c); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}(command)
	}
	wg.Wait()
	return firstErr
}

func (s *Shell) executeCmd(cmd Command) error {
	if handler, ok := s.handlers[cmd.Name]; ok {
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
