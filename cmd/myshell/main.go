package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

func main() {
	builtins := []string{"type", "exit", "echo"}
	for {
		// Uncomment this block to pass the first stage
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			os.Exit(1)
		}
		command = strings.Trim(command, "\n")
		parts := strings.Split(command, " ")

		com := parts[0]
		args := parts[1:]

		switch com {
		case "exit":
			exitCode, err := strconv.Atoi(args[0])
			if err != nil {
				os.Exit(1)
			}
			os.Exit(exitCode)
		case "echo":
			fmt.Fprintln(os.Stdout, strings.Join(args, " "))
		case "type":
			if slices.Contains(builtins, args[0]) {
				fmt.Fprintln(os.Stdout, args[0]+" is a shell builtin")
			} else if isExecutableInPath(args[0]) != "" {
				fmt.Fprintln(os.Stdout, isExecutableInPath(args[0]))
			} else {
				fmt.Fprintln(os.Stdout, args[0]+": not found")
			}
		default:
			fmt.Fprintln(os.Stdout, command+": command not found")
		}

	}
}

func isExecutableInPath(executable string) string {
	path := os.Getenv("PATH")
	paths := strings.Split(path, ":")
	for _, path := range paths {
		exec := path + "/" + executable
		if _, err := os.Stat(exec); err == nil {

			return executable + " is " + exec
		}
	}
	return ""
}
