package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Fprint

func main() {
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
		default:
			fmt.Fprintln(os.Stdout, command+": command not found")
		}

	}
}
