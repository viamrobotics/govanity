package gen

import (
	"fmt"

	"github.com/edaniels/golog"

	"go.viam.com/govanity"
)

func VanityHost() {
	_, vanity, err := govanity.ParseModules()
	if err != nil {
		golog.Global.Fatal(err)
	}
	fmt.Print(vanity)
}
