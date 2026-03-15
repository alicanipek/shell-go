package main

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

var builtins = []string{"type", "exit", "echo", "pwd", "cd"}

func (s *Shell) dispatch(cmd string, args []string, stdout, stderr *os.File) {
	switch cmd {
	case "exit":
		code := 0
		if len(args) > 0 {
			var err error
			if code, err = strconv.Atoi(args[0]); err != nil {
				code = 1
			}
		}
		os.Exit(code)

	case "echo":
		fmt.Fprintln(stdout, strings.Join(args, " "))

	case "type":
		if len(args) == 0 {
			return
		}
		if slices.Contains(builtins, args[0]) {
			fmt.Fprintln(stdout, args[0]+" is a shell builtin")
		} else if desc := isExecutableInPath(args[0]); desc != "" {
			fmt.Fprintln(stdout, desc)
		} else {
			fmt.Fprintln(stdout, args[0]+": not found")
		}

	case "pwd":
		dir, err := os.Getwd()
		if err != nil {
			os.Exit(1)
		}
		fmt.Fprintln(stdout, dir)

	case "cd":
		if len(args) == 0 {
			return
		}
		p := args[0]
		if p == "~" {
			p, _ = os.UserHomeDir()
		}
		if err := os.Chdir(p); err != nil {
			fmt.Fprintf(stderr, "cd: %s: No such file or directory\n", args[0])
		}

	default:
		if err := execute(cmd, args, stdout, stderr); err != nil && isNotFound(err) {
			fmt.Fprintln(stderr, cmd+": command not found")
		}
	}
}
