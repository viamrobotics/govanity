package gen

import (
	"fmt"

	"go.viam.com/govanity"

	"github.com/edaniels/golog"
)

func Phony() {
	modules, err := govanity.ParseModules()
	if err != nil {
		golog.Global.Fatal(err)
	}
	fmt.Println("package doc")
	for k := range modules {
		fmt.Printf("import _ \"%s\"\n", k)
	}
}
