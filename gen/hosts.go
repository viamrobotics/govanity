package gen

import (
	"fmt"
	"strings"

	"go.viam.com/govanity"

	"github.com/edaniels/golog"
)

func Hosts() {
	modules, err := govanity.ParseModules()
	if err != nil {
		golog.Global.Fatal(err)
	}
	seen := map[string]struct{}{}
	for _, module := range modules {
		if !module.Vanity {
			continue
		}
		vanity := strings.Split(module.Name, "/")[0]
		if _, ok := seen[vanity]; ok {
			continue
		}
		seen[vanity] = struct{}{}
		fmt.Printf("127.0.0.1 %s\n", vanity)
	}
}
