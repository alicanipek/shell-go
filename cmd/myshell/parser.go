package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Command struct {
	Name   string
	Args   []string
	Stdout *os.File
	Stderr *os.File
	Stdin  *os.File
}

func parseInput(input string) []Command {
	commands := strings.Split(input, "|")
	var result []Command
	for _, cmd := range commands {

		stdout := os.Stdout
		stderr := os.Stderr
		command := cmd

		switch {
		case strings.Contains(input, "1>>"):
			parts := strings.SplitN(input, "1>>", 2)
			command = strings.TrimSpace(parts[0])
			stdout = openFile(strings.TrimSpace(parts[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY)
		case strings.Contains(input, "2>>"):
			parts := strings.SplitN(input, "2>>", 2)
			command = strings.TrimSpace(parts[0])
			stderr = openFile(strings.TrimSpace(parts[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY)
		case strings.Contains(input, "2>"):
			parts := strings.SplitN(input, "2>", 2)
			command = strings.TrimSpace(parts[0])
			stderr = openFile(strings.TrimSpace(parts[1]), os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
		case strings.Contains(input, ">>"):
			parts := strings.SplitN(input, ">>", 2)
			command = strings.TrimSpace(parts[0])
			stdout = openFile(strings.TrimSpace(parts[1]), os.O_APPEND|os.O_CREATE|os.O_WRONLY)
		case strings.Contains(input, "1>"):
			parts := strings.SplitN(input, "1>", 2)
			command = strings.TrimSpace(parts[0])
			stdout = openFile(strings.TrimSpace(parts[1]), os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
		case strings.Contains(input, ">"):
			parts := strings.SplitN(input, ">", 2)
			command = strings.TrimSpace(parts[0])
			stdout = openFile(strings.TrimSpace(parts[1]), os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
		}

		args := tokenize(command)
		result = append(
			result,
			Command{
				Name:   args[0],
				Args:   args[1:],
				Stdout: stdout,
				Stderr: stderr,
				Stdin:  os.Stdin,
			},
		)
	}
	return result
}

func tokenize(command string) []string {
	var args []string
	var token strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escapeNext := false

	for i := range len(command) {
		char := command[i]

		switch {
		case escapeNext:
			token.WriteByte(char)
			escapeNext = false

		case char == '\\' && !inSingleQuote:
			if inDoubleQuote {
				var next byte
				if i+1 < len(command) {
					next = command[i+1]
				}
				if next == '"' || next == '\\' || next == '$' || next == '`' {
					escapeNext = true
				} else {
					token.WriteByte(char)
				}
			} else {
				escapeNext = true
			}

		case char == '\'' && !inDoubleQuote:
			inSingleQuote = !inSingleQuote

		case char == '"' && !inSingleQuote:
			inDoubleQuote = !inDoubleQuote

		case char == ' ' && !inSingleQuote && !inDoubleQuote:
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
	return args
}

func openFile(path string, flag int) *os.File {
	abs, _ := filepath.Abs(path)
	if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
		fmt.Fprintln(os.Stderr, "error creating directories:", err)
		os.Exit(1)
	}
	f, err := os.OpenFile(abs, flag, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error opening file:", err)
	}
	return f
}
