package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func (s *Shell) builtinExit(cmd Command) error {
	code := 0
	if len(cmd.Args) > 0 {
		var err error
		if code, err = strconv.Atoi(cmd.Args[0]); err != nil {
			code = 1
		}
	}
	return &ExitError{code}
}

func (s *Shell) builtinEcho(cmd Command) error {
	fmt.Fprintln(cmd.Stdout, strings.Join(cmd.Args, " "))
	return nil
}

func (s *Shell) builtinType(cmd Command) error {
	if len(cmd.Args) == 0 {
		return nil
	}
	name := cmd.Args[0]
	for _, b := range builtinNames {
		if b == name {
			fmt.Fprintln(cmd.Stdout, name+" is a shell builtin")
			return nil
		}
	}
	if desc := isExecutableInPath(name); desc != "" {
		fmt.Fprintln(cmd.Stdout, desc)
	} else {
		fmt.Fprintln(cmd.Stderr, name+": not found")
	}
	return nil
}

func (s *Shell) builtinPwd(cmd Command) error {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(cmd.Stderr, "pwd: error getting current directory")
		return nil
	}
	fmt.Fprintln(cmd.Stdout, dir)
	return nil
}

func (s *Shell) builtinCd(cmd Command) error {
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

func (s *Shell) builtinHistory(cmd Command) error {
	n := len(s.history)
	if len(cmd.Args) == 1 {
		var err error
		if n, err = strconv.Atoi(cmd.Args[0]); err != nil {
			n = 0
		}
	}

	if len(cmd.Args) == 2 && cmd.Args[0] == "-r" {
		fileName := cmd.Args[1]
		historyFromFile, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Fprintf(cmd.Stderr, "history: error reading file: %v\n", err)
			return nil
		}
		lines := strings.Split(string(historyFromFile), "\n")
		for _, line := range lines {
			if line != "" {
				s.history = append(s.history, line)
			}
		}
		return nil
	}

	for i, entry := range s.history[len(s.history)-n:] {
		fmt.Fprintf(cmd.Stdout, "%d %s\n", len(s.history)-n+i+1, entry)
	}
	return nil
}
