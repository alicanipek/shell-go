package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

func execute(name string, args []string, stdout, stderr *os.File) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func isExecutableInPath(exe string) string {
	for _, dir := range strings.Split(os.Getenv("PATH"), ":") {
		full := dir + "/" + exe
		if info, err := os.Stat(full); err == nil && info.Mode()&0111 != 0 {
			return exe + " is " + full
		}
	}
	return ""
}

func isNotFound(err error) bool {
	return errors.Is(err, exec.ErrNotFound)
}
