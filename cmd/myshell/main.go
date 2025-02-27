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
	for {
		fmt.Fprint(os.Stdout, "$ ")

		input := readInput(os.Stdin)
		command := strings.Trim(string(input), "\n")

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
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer term.Restore(fd, oldState)
	reader := bufio.NewReader(rd)
	// Uncomment this block to pass the first stage

	// Wait for user input
	tabcount := 0
	for {
		b, _, err := reader.ReadRune()
		if err != nil {
			fmt.Println(err)
			continue
		}

		if b == '\x03' {
			os.Exit(0)
		} else if b == '\n' || b == '\r' {
			fmt.Fprint(os.Stdout, "\r\n")
			break
		} else if b == '\x7F' {
			if length := len(input); length > 0 {
				input = input[:length-1]
				fmt.Fprint(os.Stdout, "\b \b")
			}
		} else if b == '\t' {
			executablesInPath, _ := getExecutablesInPath()
			execs := concat(builtins, executablesInPath)
			filtered := filter(execs, input)
			slices.Sort(filtered)
			if len(filtered) > 1 {
				if tabcount == 0 {
					p_cmd := find_common_name(filtered)
					if p_cmd != input {
						input = p_cmd
						fmt.Print("\r\x1b[K")
						fmt.Printf("$ %s", input)
					} else {
						tabcount++
						fmt.Print("\a")
					}

				} else {
					term.Restore(fd, oldState)
					matching := strings.Join(filtered, "  ")
					fmt.Fprint(os.Stdout, "\r\n")
					fmt.Fprint(os.Stdout, matching)
					fmt.Fprint(os.Stdout, "\r\n")
					fmt.Fprint(os.Stdout, "$ ")
					term.MakeRaw(fd)
					fmt.Fprint(os.Stdout, input)
					tabcount = 0
				}
			} else {
				for _, v := range execs {
					splitted := strings.Split(input, " ")

					after, found := strings.CutPrefix(v, splitted[0])
					if found {
						for i := 0; i < len(splitted[1:]); i++ {
							fmt.Fprint(os.Stdout, "\b")
						}
						fmt.Fprint(os.Stdout, after+" "+strings.Join(splitted[1:], " "))
						input = v + " " + strings.Join(splitted[1:], " ")
						break
					} else {
						fmt.Fprint(os.Stdout, "\a")
					}
				}
			}
		} else {
			input += string(b)
			fmt.Fprint(os.Stdout, string(b))
		}
	}
	return
}

func concat(slice1 []string, slice2 []string) []string {
	var concatted []string
	for _, v := range slice1 {
		if slices.Contains(concatted, v) {
			continue
		}
		concatted = append(concatted, v)
	}
	for _, v := range slice2 {
		if slices.Contains(concatted, v) {
			continue
		}
		concatted = append(concatted, v)
	}
	return concatted
}

func find_common_name(cmd_list []string) string {
	common_cmd := ""
	is_common := true
	for i := 0; is_common; i++ {
		var curr_letter byte
		for j := 0; j < len(cmd_list); j++ {
			if i >= len(cmd_list[j]) {
				is_common = false
				break
			} else if j == 0 {
				curr_letter = cmd_list[j][i]
			} else if curr_letter != cmd_list[j][i] {
				is_common = false
				break
			}
		}
		if is_common {
			common_cmd += string(curr_letter)
		}
	}
	return common_cmd
}

func filter(executables []string, input string) []string {
	var filtered []string
	for _, v := range executables {
		_, found := strings.CutPrefix(v, input)
		if found {
			filtered = append(filtered, v)
		}
	}
	return filtered
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
			token.WriteByte(char)
			escapeNext = false

		case char == '\\' && !inSingleQuote:
			if inDoubleQuote {
				nextChar := byte(0)
				if i+1 < len(input) {
					nextChar = input[i+1]
				}
				if nextChar == '"' || nextChar == '\\' || nextChar == '$' || nextChar == '`' {
					escapeNext = true
				} else {
					token.WriteByte(char)
				}
			} else {
				escapeNext = true
			}

		case char == '\'' && !inDoubleQuote && !escapeNext:
			inSingleQuote = !inSingleQuote

		case char == '"' && !inSingleQuote && !escapeNext:
			inDoubleQuote = !inDoubleQuote

		case char == ' ' && !inSingleQuote && !inDoubleQuote && !escapeNext:
			if token.Len() > 0 {
				args = append(args, token.String())
				token.Reset()
			}
		default:
			token.WriteByte(char)
		}
	}

	if token.Len() > 0 {
		args = append(args, token.String())
	}

	return args, file, error
}

func createFile(path string, flag int) *os.File {
	abs, _ := filepath.Abs(path)
	parentDir := filepath.Dir(abs)
	err := os.MkdirAll(parentDir, 0755)
	if err != nil {
		fmt.Println("Error creating directories:", err)
		os.Exit(1)
	}

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

func getExecutablesInPath() ([]string, error) {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return nil, fmt.Errorf("PATH environment variable is not set")
	}

	pathDirs := filepath.SplitList(pathEnv)

	executables := make(map[string]bool)

	for _, dir := range pathDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			fullPath := filepath.Join(dir, entry.Name())

			if isExecutable(fullPath) {
				executables[entry.Name()] = true
			}
		}
	}

	var result []string
	for exe := range executables {
		result = append(result, exe)
	}

	return result, nil
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.Mode().Perm()&0111 != 0 || strings.HasSuffix(strings.ToLower(path), ".exe")
}
