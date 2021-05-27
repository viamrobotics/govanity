package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"go.viam.com/govanity/server"

	"github.com/edaniels/golog"
	"go.viam.com/utils/secrets"
)

func main() {
	debug := flag.Bool("debug", false, "debug mode")
	enableDocs := flag.Bool("enable-docs", false, "enable docs")
	staticDocs := flag.Bool("static-docs", false, "enable docs")
	port := flag.Int("port", 8080, "http port")

	flag.Parse()

	var goModFilePath string
	if *enableDocs {
		if flag.NArg() != 1 {
			fmt.Fprintln(os.Stderr, "govanity <go mod file>")
			flag.Usage()
			os.Exit(1)
		}
		goModFilePath = flag.Arg(0)
	}

	secretSource, err := secrets.NewSecretSource(
		context.Background(),
		secrets.SecretSourceType(os.Getenv("SECRET_SOURCE")))
	if err != nil {
		golog.Global.Fatal(err)
	}

	docsUsername, err := secretSource.Get(context.Background(), "DOCS_USERNAME")
	if err != nil && err != secrets.ErrSecretNotFound {
		golog.Global.Fatal(err)
	}
	docsPassword, err := secretSource.Get(context.Background(), "DOCS_PASSWORD")
	if err != nil && err != secrets.ErrSecretNotFound {
		golog.Global.Fatal(err)
	}
	docOptions := server.DocOptions{
		Enabled:     *enableDocs,
		Static:      *staticDocs,
		ModFilePath: goModFilePath,
		Username:    docsUsername,
		Password:    docsPassword,
	}

	server, err := server.NewServer(*port, docOptions, secretSource, *debug)
	if err != nil {
		golog.Global.Fatal(err)
	}
	if err := server.Run(); err != nil {
		golog.Global.Fatal(err)
	}
}
