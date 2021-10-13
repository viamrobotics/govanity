package gen

import (
	"fmt"
	"net/http"

	"github.com/edaniels/golog"

	"go.viam.com/govanity"
)

func Hosts() {
	modules, vanity, err := govanity.ParseModules()
	if err != nil {
		golog.Global.Fatal(err)
	}
	if vanity == "" {
		return
	}
	fmt.Printf("127.0.0.1 %s\n", vanity)
	for _, module := range modules {
		if module.Vanity {
			continue
		}
		resp, err := http.Get(fmt.Sprintf("https://%s", module.RedirectTo))
		if err != nil {
			golog.Global.Debugw("error doing HTTP GET", "module", module.RedirectTo, "error", err)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			continue
		}
		fmt.Printf("%s,", module.Name)
	}
}
