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
	for k, v := range modules {
		if k == v {
			// no vanity
			continue
		}
		fmt.Printf("%s,", k)
	}
}
