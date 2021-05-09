package main

import (
	"context"
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

	secretSource, err := govanity.NewSecretSource(
		context.Background(),
		govanity.SecretSourceType(os.Getenv("SECRET_SOURCE")))
	if err != nil {
		golog.Global.Fatal(err)
	}

	docsUsername, err := secretSource.Get(context.Background(), "docs_username")
	if err != nil && err != govanity.ErrSecretNotFound {
		golog.Global.Fatal(err)
	}
	docsPassword, err := secretSource.Get(context.Background(), "docs_password")
	if err != nil && err != govanity.ErrSecretNotFound {
		golog.Global.Fatal(err)
	}
	docOptions := govanity.DocOptions{
		Enabled:  *enableDocs,
		Static:   *staticDocs,
		Username: docsUsername,
		Password: docsPassword,
	}

	server, err := govanity.NewServer(*port, docOptions, secretSource, *debug)
	if err != nil {
		golog.Global.Fatal(err)
	}
	if err := server.Run(); err != nil {
		golog.Global.Fatal(err)
	}
}
