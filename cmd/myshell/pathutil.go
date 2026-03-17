package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

func getExecutablesInPath() ([]string, error) {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return nil, fmt.Errorf("PATH environment variable is not set")
	}

	var (
		mu          sync.Mutex
		executables = make(map[string]struct{})
		wg          sync.WaitGroup
	)

	for _, dir := range filepath.SplitList(pathEnv) {
		if strings.HasPrefix(dir, "/mnt/") {
			continue
		}
		wg.Add(1)
		go func(d string) {
			defer wg.Done()
			entries, err := os.ReadDir(d)
			if err != nil {
				return
			}
			var local []string
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				var perm os.FileMode
				if entry.Type().IsRegular() {
					info, err := entry.Info()
					if err != nil {
						continue
					}
					perm = info.Mode().Perm()
				} else {
					info, err := os.Stat(filepath.Join(d, entry.Name()))
					if err != nil {
						continue
					}
					perm = info.Mode().Perm()
				}
				if isExecutable(entry.Name(), perm) {
					local = append(local, entry.Name())
				}
			}
			mu.Lock()
			for _, name := range local {
				executables[name] = struct{}{}
			}
			mu.Unlock()
		}(dir)
	}
	wg.Wait()

	result := make([]string, 0, len(executables))
	for exe := range executables {
		result = append(result, exe)
	}
	return result, nil
}

func isExecutable(name string, perm os.FileMode) bool {
	return perm&0111 != 0 || strings.HasSuffix(strings.ToLower(name), ".exe")
}

func getFilesAndDirectories(folder string) ([]string, []string, error) {
	pwd := os.Getenv("PWD")
	entries, err := os.ReadDir(path.Join(pwd, folder))
	if err != nil {
		return nil, nil, err
	}
	var files []string
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		} else {
			files = append(files, entry.Name())
		}
	}
	return files, dirs, nil
}

func concat(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	result := make([]string, 0, len(a)+len(b))
	for _, v := range a {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	for _, v := range b {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

func filter(candidates []string, input string) []string {
	var result []string
	for _, v := range candidates {
		if _, ok := strings.CutPrefix(v, input); ok {
			result = append(result, v)
		}
	}
	return result
}

func findCommonPrefix(list []string) string {
	if len(list) == 0 {
		return ""
	}
	var b strings.Builder
	for i := range len(list[0]) {
		ch := list[0][i]
		for _, s := range list[1:] {
			if i >= len(s) || s[i] != ch {
				return b.String()
			}
		}
		b.WriteByte(ch)
	}
	return b.String()
}
