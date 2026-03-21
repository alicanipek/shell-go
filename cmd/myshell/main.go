package main

import (
	"fmt"
	"io"
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
		execute(commands)

		// if len(commands) == 1 {
		// 	s.dispatch(commands[0])
		// 	closeIfNotStd(commands[0].Stdout)
		// 	closeIfNotStd(commands[0].Stderr)
		// 	closeIfNotStd(commands[0].Stdin)
		// } else {
		// 	c1 := exec.Command(commands[0].Name, commands[0].Args...)
		// 	c1.Stdout = commands[0].Stdout
		// 	c1.Stderr = commands[0].Stderr
		// 	c1.Stdin = commands[0].Stdin
		// 	c2 := exec.Command(commands[1].Name, commands[1].Args...)
		// 	c2.Stdout = commands[1].Stdout
		// 	c2.Stderr = commands[1].Stderr
		// 	c2.Stdin = commands[1].Stdin

		// 	pr, pw := io.Pipe()
		// 	c1.Stdout = pw
		// 	c2.Stdin = pr
		// 	c2.Stdout = os.Stdout

		// 	c1.Start()
		// 	c2.Start()
		// 	go func() {
		// 		defer pw.Close()

		// 		c1.Wait()
		// 	}()
		// 	c2.Wait()

		// }
	}
}

func closeIfNotStd(f *io.Writer) {
	if closer, ok := (*f).(io.Closer); ok {
		closer.Close()
	}

}
