package govanity

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func ResolveFile(fn string) string {
	_, thisFilePath, _, _ := runtime.Caller(0)
	thisDirPath, err := filepath.Abs(filepath.Dir(thisFilePath))
	if err != nil {
		panic(err)
	}
	return filepath.Join(thisDirPath, fn)
}

func ParseModules() ([]Module, error) {
	moduleFiles, err := os.Open(ResolveFile("./etc/modules.txt"))
	if err != nil {
		return nil, err
	}
	var modules []Module
	scanner := bufio.NewScanner(moduleFiles)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "=", 2)
		var module Module
		module.Name = parts[0]
		if len(parts) == 1 {
			module.RedirectTo = parts[0]
		} else {
			module.RedirectTo = parts[1]
			module.Vanity = true
		}
		modules = append(modules, module)
	}
	moduleFiles.Close()
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return modules, nil
}
