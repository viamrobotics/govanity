package govanity

import (
	"fmt"
	"os"
	"strings"
)

type Module struct {
	Name       string
	RedirectTo string
	Private    bool
	Vanity     bool
}

const (
	// GoModulesEnvVarName is the environment variable for specifying
	// the go modules to vanity (or provide docs for). It's
	// comma delimited.
	// e.g. vanity.domain.com/private-repo=github.com/org/private-repo,github.com/org/no-vanity-private-repro
	GoModulesEnvVarName = "GO_MODULES"
)

func ParseModules() ([]Module, string, error) {
	moduleRaw, ok := os.LookupEnv(GoModulesEnvVarName)
	if !ok || moduleRaw == "" {
		return nil, "", fmt.Errorf("at least one module must be specified in %q", GoModulesEnvVarName)
	}
	modulesRawSplit := strings.Split(moduleRaw, ",")
	var modules []Module
	var vanityHost string
	for _, moduleRaw := range modulesRawSplit {
		parts := strings.SplitN(moduleRaw, "=", 2)
		var module Module
		module.Name = parts[0]
		if len(parts) == 1 {
			module.RedirectTo = parts[0]
		} else {
			module.RedirectTo = parts[1]
			module.Vanity = true
			thisVanity := strings.Split(module.Name, "/")[0]
			if vanityHost == "" {
				vanityHost = thisVanity
			} else if vanityHost != thisVanity {
				return nil, "", fmt.Errorf("cannot have more than one vanity; found %q and %q", vanityHost, thisVanity)
			}
		}
		modules = append(modules, module)
	}
	return modules, vanityHost, nil
}
