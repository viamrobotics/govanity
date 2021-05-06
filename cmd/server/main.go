package main

import (
	"flag"
	"os"

	"go.viam.com/govanity"

	"github.com/edaniels/golog"
)

func main() {
	debug := flag.Bool("debug", false, "debug mode")
	enableDocs := flag.Bool("enable-docs", true, "enable docs")
	staticDocs := flag.Bool("static-docs", false, "enable docs")
	port := flag.Int("port", 8080, "http port")

	flag.Parse()

	server, err := govanity.NewServer(
		*port,
		govanity.DocOptions{
			Enabled:  *enableDocs,
			Static:   *staticDocs,
			Username: os.Getenv("DOCS_USERNAME"),
			Password: os.Getenv("DOCS_PASSWORD"),
		},
		os.Getenv("USE_GCP_SECRETS") != "",
		*debug)
	if err != nil {
		golog.Global.Fatal(err)
	}
	if err := server.Run(); err != nil {
		golog.Global.Fatal(err)
	}
}
