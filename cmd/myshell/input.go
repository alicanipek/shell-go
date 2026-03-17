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

	allMatches := append(append([]string{}, dirMatches...), matches...)
	slices.Sort(allMatches)
	if len(allMatches) == 0 {
		fmt.Print("\a")
		return
	}

	// Single exact match: append trailing separator
	if len(allMatches) == 1 {
		suffix := " "
		if len(matches) == 0 { // directory
			suffix = "/"
		}
		*input = cmdPart + folderPrefix + allMatches[0] + suffix
		fmt.Printf("\r\x1b[K$ %s", *input)
		*tabCount = 0
		return
	}

	// Always complete to LCP immediately (resets tabCount if progress was made)
	if common := findCommonPrefix(allMatches); common != filePart {
		*input = cmdPart + folderPrefix + common
		fmt.Printf("\r\x1b[K$ %s", *input)
		*tabCount = 0
		return
	}

	// Multiple matches with no further common prefix
	if *tabCount == 0 {
		fmt.Print("\a")
		(*tabCount)++
		return
	}

	// Second tab: list all candidates (dirs shown with trailing /)
	display := make([]string, 0, len(allMatches))
	for _, d := range dirMatches {
		display = append(display, d+"/")
	}
	display = append(display, matches...)
	slices.Sort(display)
	term.Restore(fd, oldState)
	fmt.Fprintf(os.Stdout, "\r\n%s\r\n$ ", strings.Join(display, "  "))
	term.MakeRaw(fd)
	fmt.Fprint(os.Stdout, *input)
	*tabCount = 0
}

func (s *Shell) completeCommand(fd int, oldState *term.State, input *string, tabCount *int) {
	matches := filter(s.allCommands, *input)
	slices.Sort(matches)

	if len(matches) == 0 {
		fmt.Print("\a")
		return
	}

	// Single exact match: append space
	if len(matches) == 1 {
		*input = matches[0] + " "
		fmt.Printf("\r\x1b[K$ %s", *input)
		*tabCount = 0
		return
	}

	// Always complete to LCP immediately
	if common := findCommonPrefix(matches); common != *input {
		*input = common
		fmt.Printf("\r\x1b[K$ %s", *input)
		*tabCount = 0
		return
	}

	// Multiple matches with no further common prefix
	if *tabCount == 0 {
		fmt.Print("\a")
		(*tabCount)++
		return
	}

	// Second tab: list all candidates
	term.Restore(fd, oldState)
	fmt.Fprintf(os.Stdout, "\r\n%s\r\n$ ", strings.Join(matches, "  "))
	term.MakeRaw(fd)
	fmt.Fprint(os.Stdout, *input)
	*tabCount = 0
}
