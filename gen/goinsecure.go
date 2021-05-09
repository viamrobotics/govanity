package gen

import (
	"fmt"

	"go.viam.com/govanity"

	"github.com/edaniels/golog"
)

func GoInsecure() {
	modules, err := govanity.ParseModules()
	if err != nil {
		golog.Global.Fatal(err)
	}
	for _, module := range modules {
		if !module.Vanity {
			continue
		}
		fmt.Printf("%s,", module.Name)
	}
}
