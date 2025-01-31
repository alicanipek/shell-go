package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
)

func main() {
	builtins := []string{"type", "exit", "echo", "pwd"}
	for {
		// Uncomment this block to pass the first stage
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			os.Exit(1)
		}
		command = strings.Trim(command, "\n")
		parts := parseInput(command)

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
			} else {
				inPath := isExecutableInPath(args[0])
				if inPath != "" {
					fmt.Fprintln(os.Stdout, inPath)
				} else {
					fmt.Fprintln(os.Stdout, args[0]+": not found")
				}
			}
		case "pwd":
			dir, err := os.Getwd()
			if err != nil {
				os.Exit(1)
			}
			fmt.Fprintln(os.Stdout, dir)
		case "cd":
			var p string
			if args[0] == "~" {
				p, _ = os.UserHomeDir()
			} else {
				p = args[0]
			}
			err := os.Chdir(p)
			if err != nil {
				fmt.Println("cd: " + args[0] + ": No such file or directory")
			}
		default:
			err := execute(com, args)
			if err != nil {
				fmt.Fprintln(os.Stdout, command+": command not found")
			}
		}

	}
}

func execute(executable string, args []string) error {

	cmd := exec.Command(executable, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	er := cmd.Run()
	return er
}

func parseInput(input string) []string {
	var args []string
	var token strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escapeNext := false

	for i := 0; i < len(input); i++ {
		char := input[i]

		switch {
		case escapeNext:
			// Handle escaped character (treat it as a literal)
			token.WriteByte(char)
			escapeNext = false

		case char == '\\' && !inSingleQuote:
			// Escape the next character (only if not inside single quotes)
			escapeNext = true

		case char == '\'' && !inDoubleQuote && !escapeNext:
			// Toggle single quote state (only if not inside double quotes and not escaped)
			inSingleQuote = !inSingleQuote

		case char == '"' && !inSingleQuote && !escapeNext:
			// Toggle double quote state (only if not inside single quotes and not escaped)
			inDoubleQuote = !inDoubleQuote

		case char == ' ' && !inSingleQuote && !inDoubleQuote && !escapeNext:
			// End of token, add to args
			if token.Len() > 0 {
				args = append(args, token.String())
				token.Reset()
			}

		default:
			// Add character to current token
			token.WriteByte(char)
		}
	}

	// Add the last token if it exists
	if token.Len() > 0 {
		args = append(args, token.String())
	}

	return args
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
