package main

import (
	"os"

	"go.viam.com/govanity/gen"

	"github.com/edaniels/golog"
)

func main() {
	if len(os.Args) == 1 {
		golog.Global.Fatal("gen command required")
	}
	cmd := os.Args[1]
	switch cmd {
	case "goinsecure":
		gen.GoInsecure()
	case "hosts":
		gen.Hosts()
	case "vanity_host":
		gen.VanityHost()
	case "phony":
		gen.Phony()
	default:
		golog.Global.Fatalf("unknown command %q", cmd)
	}
}
