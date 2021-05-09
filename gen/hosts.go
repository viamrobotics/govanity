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
	for _, module := range modules {
		if !module.Vanity {
			continue
		}
		fmt.Printf("127.0.0.1 %s\n", strings.Split(module.Name, "/")[0])
	}
}
