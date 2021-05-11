package gen

import (
	"fmt"

	"github.com/edaniels/golog"

	"go.viam.com/govanity"
)

func Hosts() {
	_, vanity, err := govanity.ParseModules()
	if err != nil {
		golog.Global.Fatal(err)
	}
	if vanity == "" {
		return
	}
	fmt.Printf("127.0.0.1 %s\n", vanity)
}
