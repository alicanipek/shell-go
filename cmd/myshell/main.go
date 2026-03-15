package main

import (
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

		args, stdout, stderr := parseInput(command)
		if len(args) == 0 {
			closeIfNotStd(stdout)
			closeIfNotStd(stderr)
			continue
		}

		s.dispatch(args[0], args[1:], stdout, stderr)
		closeIfNotStd(stdout)
		closeIfNotStd(stderr)
	}
}

func closeIfNotStd(f *os.File) {
	if f != nil && f != os.Stdout && f != os.Stderr && f != os.Stdin {
		f.Close()
	}
}
