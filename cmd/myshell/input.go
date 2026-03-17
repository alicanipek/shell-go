package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"golang.org/x/term"
)

func (s *Shell) readInput(rd io.Reader) string {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		panic(err)
	}
	defer term.Restore(fd, oldState)

	reader := bufio.NewReader(rd)
	var input string
	tabCount := 0

	for {
		b, _, err := reader.ReadRune()
		if err != nil {
			fmt.Println(err)
			continue
		}

		switch b {
		case '\x03':
			os.Exit(0)

		case '\n', '\r':
			fmt.Fprint(os.Stdout, "\r\n")
			return input

		case '\x7F':
			if n := len(input); n > 0 {
				input = input[:n-1]
				fmt.Fprint(os.Stdout, "\b \b")
			}

		case '\t':
			s.cacheReady.Wait()
			s.handleTab(fd, oldState, &input, &tabCount)

		default:
			tabCount = 0
			ch := string(b)
			input += ch
			fmt.Fprint(os.Stdout, ch)
		}
	}
}

func (s *Shell) handleTab(fd int, oldState *term.State, input *string, tabCount *int) {
	if strings.Contains(*input, " ") {
		s.completeFilePath(fd, oldState, input, tabCount)
		return
	}
	s.completeCommand(fd, oldState, input, tabCount)
}

func (s *Shell) completeFilePath(fd int, oldState *term.State, input *string, tabCount *int) {
	spaceIdx := strings.LastIndex(*input, " ") + 1
	cmdPart := *input
	filePart := ""
	if spaceIdx < len(*input) {
		filePart = (*input)[spaceIdx:]
		cmdPart = (*input)[:spaceIdx]
	}

	folderIdx := strings.LastIndex(filePart, "/")
	folder := "."
	folderPrefix := ""
	if folderIdx != -1 {
		folder = filePart[:folderIdx]
		filePart = filePart[folderIdx+1:]
		folderPrefix = folder + "/"
	}

	files, dirs, _ := getFilesAndDirectories(folder)
	matches := filter(files, filePart)
	dirMatches := filter(dirs, filePart)
	slices.Sort(matches)
	slices.Sort(dirMatches)

	if *tabCount == 0 && len(matches) == 0 && len(dirMatches) == 0 {
		fmt.Print("\a")
		return
	} else if *tabCount == 0 && len(matches) == 1 && len(dirMatches) == 0 {
		*input = cmdPart + folderPrefix + matches[0] + " "
		fmt.Printf("\r\x1b[K$ %s", *input)
		*tabCount = 0
		return
	} else if *tabCount == 0 && len(matches) == 0 && len(dirMatches) == 1 {
		*input = cmdPart + folderPrefix + dirMatches[0]
		fmt.Printf("\r\x1b[K$ %s", *input)
		*tabCount = 0
		return
	} else {
		if *tabCount == 0 {
			fmt.Print("\a")
			(*tabCount)++
			return
		} else {
			*input = cmdPart + folderPrefix + filePart
			term.Restore(fd, oldState)
			dirMatchesString := strings.Join(dirMatches, " ")
			fileMatchesString := strings.Join(matches, "  ")
			fmt.Fprintf(os.Stdout, "\r\n%s\r\n$ ", strings.TrimSpace(dirMatchesString+" "+fileMatchesString))
			term.MakeRaw(fd)
			fmt.Fprint(os.Stdout, *input)
			*tabCount = 0
			return
		}

	}

}

func (s *Shell) completeCommand(fd int, oldState *term.State, input *string, tabCount *int) {
	matches := filter(s.allCommands, *input)
	slices.Sort(matches)

	if len(matches) > 1 {
		if *tabCount == 0 {
			common := findCommonPrefix(matches)
			if common != *input {
				*input = common
				fmt.Print("\r\x1b[K")
				fmt.Printf("$ %s", *input)
			} else {
				(*tabCount)++
				fmt.Print("\a")
			}
		} else {
			term.Restore(fd, oldState)
			fmt.Fprintf(os.Stdout, "\r\n%s\r\n$ ", strings.Join(matches, "  "))
			term.MakeRaw(fd)
			fmt.Fprint(os.Stdout, *input)
			*tabCount = 0
		}
		return
	}

	matched := false
	for _, v := range s.allCommands {
		parts := strings.Split(*input, " ")
		if after, ok := strings.CutPrefix(v, parts[0]); ok {
			for range parts[1:] {
				fmt.Fprint(os.Stdout, "\b")
			}
			fmt.Fprint(os.Stdout, after+" "+strings.Join(parts[1:], " "))
			*input = v + " " + strings.Join(parts[1:], " ")
			matched = true
			break
		}
	}
	if !matched {
		fmt.Fprint(os.Stdout, "\a")
	}
}
