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
	for k, v := range modules {
		if k == v {
			// no vanity
			continue
		}
		fmt.Printf("127.0.0.1 %s\n", strings.Split(k, "/")[0])
	}
}
