package gen

import (
	"fmt"

	"github.com/edaniels/golog"

	"go.viam.com/govanity"
)

func Phony() {
	mods, _, err := govanity.ParseModules()
	if err != nil {
		golog.Global.Fatal(err)
	}
	fmt.Println("package doc")
	for _, module := range mods {
		fmt.Printf("import _ \"%s\"\n", module.Name)
	}
}
