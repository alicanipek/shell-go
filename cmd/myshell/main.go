package main

import (
	"bufio"
	"fmt"
	"os"
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
		command = strings.Trim(command, "\n")
		if err != nil {
			os.Exit(1)
		}
		if command == "exit 0" {
			os.Exit(0)
		}

		fmt.Fprint(os.Stdout, command+": command not found\n")
	}
}
