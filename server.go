package main

import (
	"context"
	"flag"
	//"fmt"
	//"io"
	//"io/ioutil"
	"net/http"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/trace"

	"github.com/edaniels/golog"
	
	"github.com/erh/egoutil"
)

var (
	repos = []string{"robotcore", "dynamixel"}
)

type MyApp struct {
	app *egoutil.SimpleWebApp
}

func (a *MyApp) init() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // there are a bunch of requests, so 30 seconds seems fair
	defer cancel()

	secrets, err := egoutil.NewGCPSecrets(ctx)
	if err != nil {
		return err
	}
	
	x := egoutil.NewSimpleWebAppConfig()

	if os.Getenv("webroot") != "" {
		x = x.SetWebRoot(os.Getenv("webroot"))
	}

	x = x.SetMongoURL(secrets.GetSecretOrPanic(ctx, "mongourl"))

	// ----

	a.app, err = egoutil.NewSimpleWebApp(ctx, x)
	if err != nil {
		return err
	}

	return nil
}

func main() {

	debug := flag.Bool("debug", false, "debug mode")

	flag.Parse()

	var err error
	var exp trace.Exporter

	if os.Getenv("webroot") == "" {
		exp = egoutil.NewNiceLoggingSpanExporter()
	} else {
		exp, err = stackdriver.NewExporter(stackdriver.Options{})
		if err != nil {
			golog.Global.Fatal(err)
		}
	}
	trace.RegisterExporter(exp)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	theApp := &MyApp{}
	err = theApp.init()
	if err != nil {
		golog.Global.Fatal(err)
	}

	if *debug {
		go theApp.app.ReloadTemplateThread()
	}

	for _, r := range repos {
		theApp.app.Mux.Handle("/" + r, &egoutil.WrappedTemplate{theApp.app, &RepoRedirct{theApp, r}, false})
	}
	theApp.app.Mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	theApp.app.Mux.Handle("/", &egoutil.WrappedTemplate{theApp.app, &IndexPage{theApp}, false})

	golog.Global.Debugf("starting to listen")

	err = http.ListenAndServe(":8080", theApp.app.Mux)
	if err != http.ErrServerClosed && err != nil {
		golog.Global.Fatal(err)
	}

}

// ----

type IndexPage struct {
	a *MyApp
}

func (p *IndexPage) Serve(ctx context.Context, user egoutil.UserInfo, r *http.Request) (string, interface{}, error) {
	type Temp struct {
		Repos []string
	}
	temp := Temp{repos}

	return "index.html", temp, nil
}

// ---

type RepoRedirct struct {
	a *MyApp
	repo string
}

func (p *RepoRedirct) Serve(ctx context.Context, user egoutil.UserInfo, r *http.Request) (string, interface{}, error) {
	type Temp struct {
		Repo string
	}
	temp := Temp{p.repo}

	return "repo.html", temp, nil
}



