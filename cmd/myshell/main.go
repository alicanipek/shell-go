package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

type Shell struct {
	cachedExecutables []string
	allCommands       []string
	cacheReady        sync.WaitGroup
}

func NewShell() *Shell {
	s := &Shell{}
	s.cacheReady.Add(1)
	go func() {
		defer s.cacheReady.Done()
		s.cachedExecutables, _ = getExecutablesInPath()
		s.allCommands = concat(builtins, s.cachedExecutables)
	}()
	return s
}

func main() {
	NewShell().run()
}

func (s *Shell) run() {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		raw := s.readInput(os.Stdin)
		command := strings.TrimRight(raw, "\n")

		commands := parseInput(command)
		if len(commands) == 0 {
			continue
		}
		if err := execute(commands); err != nil {
			var exitErr *ExitError
			if errors.As(err, &exitErr) {
				os.Exit(exitErr.Code)
			}
		}
	}
}
