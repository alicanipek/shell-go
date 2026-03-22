package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Command struct {
	Name   string
	Args   []string
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

func NewCommand() Command {
	return Command{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
	}
}

func parseInput(input string) []Command {
	parts := strings.Split(input, "|")
	var commands []Command
	for _, command := range parts {
		cmd := NewCommand()
		args := tokenize(command)
		if len(args) == 0 {
			continue
		}
		cmd.Name = args[0]
		cmd.Args = args[1:]
		commands = append(commands, cmd)
	}

	for i := 0; i < len(commands)-1; i++ {
		pipeReader, pipeWriter := io.Pipe()
		commands[i].Stdout = pipeWriter
		commands[i+1].Stdin = pipeReader
	}

	for i := range parts {
		part := parts[i]
		var target *io.Writer
		var filename string

		switch {
		case strings.Contains(part, "1>>"):
			splitParts := strings.SplitN(part, "1>>", 2)
			target = &commands[i].Stdout
			filename = strings.TrimSpace(splitParts[1])
			commands[i].Name, commands[i].Args = parseCommandWithoutRedirection(splitParts[0])
			*target = openFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY)
		case strings.Contains(part, "2>>"):
			splitParts := strings.SplitN(part, "2>>", 2)
			target = &commands[i].Stderr
			filename = strings.TrimSpace(splitParts[1])
			commands[i].Name, commands[i].Args = parseCommandWithoutRedirection(splitParts[0])
			*target = openFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY)
		case strings.Contains(part, "2>"):
			splitParts := strings.SplitN(part, "2>", 2)
			target = &commands[i].Stderr
			filename = strings.TrimSpace(splitParts[1])
			commands[i].Name, commands[i].Args = parseCommandWithoutRedirection(splitParts[0])
			*target = openFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
		case strings.Contains(part, ">>"):
			splitParts := strings.SplitN(part, ">>", 2)
			target = &commands[i].Stdout
			filename = strings.TrimSpace(splitParts[1])
			commands[i].Name, commands[i].Args = parseCommandWithoutRedirection(splitParts[0])
			*target = openFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY)
		case strings.Contains(part, "1>"):
			splitParts := strings.SplitN(part, "1>", 2)
			target = &commands[i].Stdout
			filename = strings.TrimSpace(splitParts[1])
			commands[i].Name, commands[i].Args = parseCommandWithoutRedirection(splitParts[0])
			*target = openFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
		case strings.Contains(part, ">"):
			splitParts := strings.SplitN(part, ">", 2)
			target = &commands[i].Stdout
			filename = strings.TrimSpace(splitParts[1])
			commands[i].Name, commands[i].Args = parseCommandWithoutRedirection(splitParts[0])
			*target = openFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC)
		}
	}
	return commands
}

func parseCommandWithoutRedirection(cmdStr string) (string, []string) {
	args := tokenize(cmdStr)
	if len(args) == 0 {
		return "", nil
	}
	return args[0], args[1:]
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
