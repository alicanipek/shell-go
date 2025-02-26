package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/term"
)

var builtins = []string{"type", "exit", "echo", "pwd"}

func main() {

	// commands := [3]string{
	// 	"echo 'Hello Maria' 1>> /tmp/foo/baz.md",
	// 	"echo 'Hello Emily' 1>> /tmp/foo/baz.md",
	// 	"cat /tmp/foo/baz.md"}
	// for i := 0; i < len(commands); i++ {
	for {
		fmt.Fprint(os.Stdout, "$ ")

		input := readInput(os.Stdin)
		command := strings.Trim(string(input), "\n")

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

func readInput(rd io.Reader) (input string) {

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	reader := bufio.NewReader(rd)
	// Uncomment this block to pass the first stage

	// Wait for user input

	for {
		b, _, err := reader.ReadRune()
		if err != nil {
			fmt.Println(err)
			continue
		}

		if b == '\x03' { // Ctrl+C
			os.Exit(0)
		} else if b == '\n' || b == '\r' {
			fmt.Fprint(os.Stdout, "\r\n")
			break
		} else if b == '\x7F' { // Backspace
			if length := len(input); length > 0 {
				input = input[:length-1]
				fmt.Fprint(os.Stdout, "\b \b")
			}
		} else if b == '\t' {
			for _, v := range builtins {
				after, found := strings.CutPrefix(v, input)
				if found {
					fmt.Fprint(os.Stdout, after+" ")
					input = v
					break
				}
			}

		} else {
			input += string(b)
			fmt.Fprint(os.Stdout, string(b))
		}
	}
	return
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
	if strings.Contains(input, "1>>") {
		splitInput = strings.Split(input, "1>>")
		file = createFile(strings.TrimSpace(splitInput[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY)
		error = os.Stderr
		command = strings.TrimSpace(splitInput[0])
	} else if strings.Contains(input, "2>>") {
		splitInput = strings.Split(input, "2>>")
		e := createFile(strings.TrimSpace(splitInput[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY)
		error = e
		file = os.Stdout
		command = strings.TrimSpace(splitInput[0])
	} else if strings.Contains(input, "2>") {
		splitInput = strings.Split(input, "2>")
		e := createFile(strings.TrimSpace(splitInput[1]), os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
		error = e
		file = os.Stdout
		command = strings.TrimSpace(splitInput[0])
	} else if strings.Contains(input, ">>") {
		splitInput = strings.Split(input, ">>")
		file = createFile(strings.TrimSpace(splitInput[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY)
		error = os.Stderr
		command = strings.TrimSpace(splitInput[0])
	} else if strings.Contains(input, "1>") {
		splitInput = strings.Split(input, "1>")
		file = createFile(strings.TrimSpace(splitInput[1]), os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
		error = os.Stderr
		command = strings.TrimSpace(splitInput[0])
	} else if strings.Contains(input, ">") {
		splitInput = strings.Split(input, ">")
		file = createFile(strings.TrimSpace(splitInput[1]), os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
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

func createFile(path string, flag int) *os.File {
	abs, _ := filepath.Abs(path)
	parentDir := filepath.Dir(abs)
	err := os.MkdirAll(parentDir, 0755) // 0755 is the permission mode (rwxr-xr-x)
	if err != nil {
		fmt.Println("Error creating directories:", err)
		os.Exit(1)
	}

	// Create the file
	e, err := os.OpenFile(abs, flag, 0666)
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
