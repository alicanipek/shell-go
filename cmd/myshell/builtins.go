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
	for i, entry := range s.history {
		fmt.Fprintf(cmd.Stdout, "%d %s\n", i+1, entry)
	}
	return nil
}
