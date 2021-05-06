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

func ParseModules() (map[string]string, error) {
	moduleFiles, err := os.Open(ResolveFile("./etc/modules.txt"))
	if err != nil {
		return nil, err
	}
	modules := map[string]string{}
	scanner := bufio.NewScanner(moduleFiles)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "=", 2)
		if len(parts) == 1 {
			// no vanity
			modules[parts[0]] = parts[0]
			continue
		}
		modules[parts[0]] = parts[1]
	}
	moduleFiles.Close()
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return modules, nil
}
