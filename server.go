package govanity

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
	"go.opencensus.io/trace"

	"go.viam.com/govanity/doc"

	"github.com/edaniels/golog"
	"github.com/erh/egoutil"
)

type Server struct {
	app   *MyApp
	port  int
	debug bool
}

type DocOptions struct {
	Enabled  bool
	Static   bool
	Username string
	Password string
}

type Module struct {
	Name       string
	RedirectTo string
	Private    bool
	Vanity     bool
}

func NewServer(port int, docOpts DocOptions, useGCPSecrets, debug bool) (*Server, error) {
	modules, err := ParseModules()
	if err != nil {
		return nil, err
	}

	for idx, module := range modules {
		resp, err := http.Get(fmt.Sprintf("https://%s", module.RedirectTo))
		if err != nil {
			golog.Global.Debugw("error doing HTTP GET for module %q", module.RedirectTo, "error", err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			module.Private = true
			modules[idx] = module
		}
	}

	var exp trace.Exporter

	// TODO(erd): make better way to optionally use this
	if useGCPSecrets {
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
	err = theApp.init(useGCPSecrets)
	if err != nil {
		return nil, err
	}

	goModPath := ResolveFile("doc/mod/go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		goModPath = ""
	} else {
		goModPath = filepath.Dir(goModPath)
	}
	if docOpts.Enabled {
		if goModPath == "" {
			docOpts.Enabled = false
			golog.Global.Info("No go mod found; skipping serving documentation")
		} else {
			if err := theApp.registerDocs(goModPath, docOpts); err != nil {
				return nil, err
			}
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
	if srv.debug {
		go srv.app.app.ReloadTemplateThread()
	}

	golog.Global.Debugf("starting to listen")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", srv.port), srv.app.app.Mux); err != http.ErrServerClosed && err != nil {
		golog.Global.Fatal(err)
	}
	golog.Global.Info("shutting down")
	return nil
}

type MyApp struct {
	app     *egoutil.SimpleWebApp
	modules []Module
}

func (a *MyApp) init(useGCPSecrets bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // there are a bunch of requests, so 30 seconds seems fair
	defer cancel()

	x := egoutil.NewSimpleWebAppConfig()

	if os.Getenv("webroot") != "" {
		x = x.SetWebRoot(os.Getenv("webroot"))
	}

	x.SetTemplateDir(ResolveFile("./templates"))

	// TODO(erd): make better way to optionally use this
	if useGCPSecrets {
		secrets, err := egoutil.NewGCPSecrets(ctx)
		if err != nil {
			golog.Global.Fatalw("failed to get GCP secrets", "error", err)
		}
		if url, err := secrets.GetSecret(ctx, "mongourl"); err == nil && url != "" {
			x = x.SetMongoURL(url)
		}
	}

	var err error
	a.app, err = egoutil.NewSimpleWebApp(ctx, x)
	if err != nil {
		return err
	}

	return nil
}

func (a *MyApp) registerDocs(goModPath string, opts DocOptions) error {
	absPath, err := filepath.Abs(goModPath)
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
		Modules     []Module
		DocsEnabled bool
	}
	temp := Temp{p.a.modules, p.docsEnabled}

	return "index.html", temp, nil
}

// ---

type ModuleRedirect struct {
	a      *MyApp
	module Module
}

func (p *ModuleRedirect) Serve(ctx context.Context, user egoutil.UserInfo, r *http.Request) (string, interface{}, error) {
	type Temp struct {
		Module Module
	}
	temp := Temp{p.module}

	return "module.html", temp, nil
}
