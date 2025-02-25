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

		parts, file := parseInput(command)

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
		case "cat":
			execute(com, args, file)
		default:
			err := execute(com, args, file)
			if err != nil {
				fmt.Fprintln(os.Stdout, command+": command not found")
			}
		}
		if file != nil && file != os.Stdout {
			file.Close()
		}
		file = os.Stdout
	}
}

func execute(executable string, args []string, file *os.File) error {
	cmd := exec.Command(executable, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = file
	er := cmd.Run()

	return er
}

func parseInput(input string) ([]string, *os.File) {
	var args []string
	var file *os.File
	command := input
	var token strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escapeNext := false
	var splitInput []string
	if strings.Contains(input, "1>") {
		splitInput = strings.Split(input, "1>")
	} else if strings.Contains(input, ">") {
		splitInput = strings.Split(input, ">")
	}
	if len(splitInput) > 1 {
		// abs, _ := filepath.Abs(strings.TrimSpace(splitInput[1]))
		// parentDir := filepath.Dir(abs)
		// err := os.MkdirAll(parentDir, 0755) // 0755 is the permission mode (rwxr-xr-x)
		// if err != nil {
		// 	fmt.Println("Error creating directories:", err)
		// 	os.Exit(1)
		// }

		// // Create the file
		// f, err := os.Create(abs)
		// if err != nil {
		// 	fmt.Println("Error creating files:", err)
		// 	os.Exit(1)
		// }
		file, _ = os.OpenFile(strings.TrimSpace(splitInput[1]), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		command = strings.TrimSpace(splitInput[0])
	} else {
		file = os.Stdout
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

	return args, file
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
