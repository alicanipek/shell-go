package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

var builtinNames = []string{"type", "exit", "echo", "pwd", "cd", "history"}

type Shell struct {
	history             []string
	historyAppendOffset int
	cachedExecutables   []string
	allCommands         []string
	cacheReady          sync.WaitGroup
	handlers            map[string]func(Command) error
}

func NewShell() *Shell {
	s := &Shell{}
	s.handlers = map[string]func(Command) error{
		"exit":    s.builtinExit,
		"echo":    s.builtinEcho,
		"type":    s.builtinType,
		"pwd":     s.builtinPwd,
		"cd":      s.builtinCd,
		"history": s.builtinHistory,
	}
	s.cacheReady.Add(1)
	go func() {
		defer s.cacheReady.Done()
		s.cachedExecutables, _ = getExecutablesInPath()
		s.allCommands = concat(builtinNames, s.cachedExecutables)
	}()

	s.history = loadHistory()
	s.historyAppendOffset = len(s.history)
	return s
}

func loadHistory() []string {
	historyFile := os.Getenv("HISTFILE")
	if historyFile == "" {
		return nil
	}
	f, err := os.Open(historyFile)
	if err != nil {
		return nil
	}
	defer f.Close()
	var history []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			history = append(history, line)
		}
	}
	return history
}

func (s *Shell) run() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		raw := s.readInput(os.Stdin)
		command := strings.TrimRight(raw, "\n")
		s.history = append(s.history, command)

		commands := parseInput(command)
		if len(commands) == 0 {
			continue
		}
		if err := s.execute(commands); err != nil {
			var exitErr *ExitError
			if errors.As(err, &exitErr) {
				os.Exit(exitErr.Code)
			}
		}
	}
}
