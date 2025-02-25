package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

func main() {
	builtins := []string{"type", "exit", "echo", "pwd"}

	// commands := [2]string{"ls -1 nonexistent 2> /tmp/qux/bar.md", "cat /tmp/qux/bar.md"}
	// for i := 0; i < len(commands); i++ {
	for {
		// Uncomment this block to pass the first stage
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			os.Exit(1)
		}
		command = strings.Trim(command, "\n")

		// command := commands[i]
		parts, file, error := parseInput(command)

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
			fmt.Fprintln(file, strings.Join(args, " "))
		case "type":
			if slices.Contains(builtins, args[0]) {
				fmt.Fprintln(file, args[0]+" is a shell builtin")
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
		case "cat", "ls":
			execute(com, args, file, error)
		default:
			err := execute(com, args, file, error)
			if err != nil {
				fmt.Fprintln(error, command+": command not found")
			}
		}
		if file != nil && file != os.Stdout {
			file.Close()
		}
		file = os.Stdout
		if error != nil && error != os.Stderr {
			error.Close()
		}
		error = os.Stderr
	}
}

func execute(executable string, args []string, file *os.File, error *os.File) error {
	cmd := exec.Command(executable, args...)
	cmd.Stderr = error
	cmd.Stdout = file
	er := cmd.Run()
	return er
}

func parseInput(input string) ([]string, *os.File, *os.File) {
	var args []string
	var file *os.File
	var error *os.File
	command := input
	var token strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escapeNext := false
	var splitInput []string
	if strings.Contains(input, "1>") {
		splitInput = strings.Split(input, "1>")
		file = createFile(strings.TrimSpace(splitInput[1]))
		error = os.Stderr
		command = strings.TrimSpace(splitInput[0])
	} else if strings.Contains(input, "2>") {
		splitInput = strings.Split(input, "2>")
		e := createFile(strings.TrimSpace(splitInput[1]))
		error = e
		file = os.Stdout
		command = strings.TrimSpace(splitInput[0])
	} else if strings.Contains(input, ">") {
		splitInput = strings.Split(input, ">")
		file = createFile(strings.TrimSpace(splitInput[1]))
		error = os.Stderr
		command = strings.TrimSpace(splitInput[0])
	} else {
		file = os.Stdout
		error = os.Stderr
	}
	for i := range len(command) {
		char := input[i]

		switch {
		case escapeNext:
			// Handle escaped character (treat it as a literal)
			token.WriteByte(char)
			escapeNext = false

		case char == '\\' && !inSingleQuote:
			// Escape the next character (only if not inside single quotes)
			if inDoubleQuote {
				// Inside double quotes, only escape ", $, \, and `
				nextChar := byte(0)
				if i+1 < len(input) {
					nextChar = input[i+1]
				}
				if nextChar == '"' || nextChar == '\\' || nextChar == '$' || nextChar == '`' {
					escapeNext = true
				} else {
					// Treat the backslash as a literal
					token.WriteByte(char)
				}
			} else {
				// Outside quotes, escape the next character
				escapeNext = true
			}

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

	return args, file, error
}

func createFile(path string) *os.File {
	abs, _ := filepath.Abs(path)
	parentDir := filepath.Dir(abs)
	err := os.MkdirAll(parentDir, 0755) // 0755 is the permission mode (rwxr-xr-x)
	if err != nil {
		fmt.Println("Error creating directories:", err)
		os.Exit(1)
	}

	// Create the file
	e, err := os.OpenFile(abs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	return e
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
