package server

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/edaniels/golog"
	"github.com/erh/egoutil"
	"go.opencensus.io/trace"

	"go.viam.com/govanity"
	"go.viam.com/govanity/doc"
)

type Server struct {
	app   *MyApp
	port  int
	debug bool
}

type DocOptions struct {
	Enabled     bool
	Static      bool
	ModFilePath string
	Username    string
	Password    string
}

func NewServer(port int, docOpts DocOptions, secretSource govanity.SecretSource, debug bool) (*Server, error) {
	modules, _, err := govanity.ParseModules()
	if err != nil {
		return nil, err
	}
	for idx, module := range modules {
		resp, err := http.Get(fmt.Sprintf("https://%s", module.RedirectTo))
		if err != nil {
			golog.Global.Debugw("error doing HTTP GET", "module", module.RedirectTo, "error", err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			module.Private = true
			modules[idx] = module
		}
	}

	var exp trace.Exporter

	if secretSource.Type() == govanity.SecretSourceTypeGCP {
		// This only works with GCP right now
		var err error
		exp, err = stackdriver.NewExporter(stackdriver.Options{})
		if err != nil {
			return nil, err
		}
	} else {
		exp = egoutil.NewNiceLoggingSpanExporter()
	}
	trace.RegisterExporter(exp)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	theApp := &MyApp{modules: modules}
	if err := theApp.init(secretSource, debug); err != nil {
		return nil, err
	}

	if docOpts.Enabled {
		if err := theApp.registerDocs(docOpts); err != nil {
			return nil, err
		}
	}

	for _, module := range modules {
		if !module.Vanity {
			continue
		}
		pkgName := strings.SplitN(module.Name, "/", 2)[1]
		theApp.app.Mux.Handle("/"+pkgName, &egoutil.WrappedTemplate{theApp.app, &ModuleRedirect{theApp, module}, false})
	}
	theApp.app.Mux.Handle("/", &egoutil.WrappedTemplate{theApp.app, &IndexPage{theApp, docOpts.Enabled}, false})

	return &Server{theApp, port, debug}, nil
}

func (srv *Server) Run() error {
	golog.Global.Debugf("starting to listen")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", srv.port), srv.app.app.Mux); err != http.ErrServerClosed && err != nil {
		golog.Global.Fatal(err)
	}
	golog.Global.Info("shutting down")
	return nil
}

type MyApp struct {
	app     *egoutil.SimpleWebApp
	modules []govanity.Module
}

func (a *MyApp) init(secretSource govanity.SecretSource, debug bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // there are a bunch of requests, so 30 seconds seems fair
	defer cancel()

	var x *egoutil.SimpleWebAppConfig
	if debug {
		x = egoutil.NewSimpleWebAppConfig(govanity.ResolveFile("templates"))
	} else {
		x = egoutil.NewSimpleWebAppConfig("templates")
		x = x.SetTemplateEmbed(&govanity.EmbeddedTemplates)
	}

	if os.Getenv("webroot") != "" {
		x = x.SetWebRoot(os.Getenv("webroot"))
	}

	if url, err := secretSource.Get(ctx, "mongourl"); err == nil && url != "" {
		x = x.SetMongoURL(url)
	}

	var err error
	a.app, err = egoutil.NewSimpleWebApp(ctx, x)
	if err != nil {
		return err
	}

	return nil
}

func (a *MyApp) registerDocs(opts DocOptions) error {
	absPath, err := filepath.Abs(opts.ModFilePath)
	if err != nil {
		return err
	}
	mux, err := doc.Handler(absPath, opts.Static)
	if err != nil {
		return err
	}

	a.app.Mux.Handle("/pkg/", basicAuth(mux, opts))
	a.app.Mux.Handle("/src/", basicAuth(mux, opts))
	a.app.Mux.Handle("/cmd/", basicAuth(mux, opts))
	a.app.Mux.Handle("/lib/godoc/", basicAuth(mux, opts))
	return nil
}

// https://stackoverflow.com/questions/21936332/idiomatic-way-of-requiring-http-basic-auth-in-go
func basicAuth(handler http.Handler, opts DocOptions) http.Handler {
	if opts.Username == "" && opts.Password == "" {
		return handler
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(opts.Username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(opts.Password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="auth-zone"`)
			w.WriteHeader(401)
			if _, err := w.Write([]byte("Unauthorized.\n")); err != nil {
				golog.Global.Errorw("failed to write", "error", err)
			}
			return
		}

		handler.ServeHTTP(w, r)
	})
}

// ----

type IndexPage struct {
	a           *MyApp
	docsEnabled bool
}

func (p *IndexPage) Serve(ctx context.Context, user egoutil.UserInfo, r *http.Request) (string, interface{}, error) {
	type Temp struct {
		Modules     []govanity.Module
		DocsEnabled bool
	}
	temp := Temp{p.a.modules, p.docsEnabled}

	return "index.html", temp, nil
}

// ---

type ModuleRedirect struct {
	a      *MyApp
	module govanity.Module
}

func (p *ModuleRedirect) Serve(ctx context.Context, user egoutil.UserInfo, r *http.Request) (string, interface{}, error) {
	type Temp struct {
		Module govanity.Module
	}
	temp := Temp{p.module}

	return "module.html", temp, nil
}
