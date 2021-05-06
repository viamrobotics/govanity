package doc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"golang.org/x/tools/godoc"
	"golang.org/x/tools/godoc/static"
	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/gatefs"
	"golang.org/x/tools/godoc/vfs/mapfs"
	"golang.org/x/xerrors"
)

// adapted from https://github.com/golang/tools/

func newPresentation(goModFile string) (*godoc.Presentation, []mod, error) {
	fs := vfs.NameSpace{}

	fsGate := make(chan bool, 20)
	fs.Bind("/", vfs.NewNameSpace(), "/", vfs.BindReplace)
	fs.Bind("/lib/godoc", mapfs.New(static.Files), "/", vfs.BindReplace)
	fs.Bind("/favicon.ico", mapfs.New(static.Files), "/favicon.ico", vfs.BindReplace)

	// Determine modules in the build list.
	mods, err := buildList(goModFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to determine the build list of the main module: %w", err)
	}

	// Bind module trees into Go root.
	for _, m := range mods {
		if m.Dir == "" {
			// Module is not available in the module cache, skip it.
			continue
		}
		dst := path.Join("/src", m.Path)
		fs.Bind(dst, gatefs.New(vfs.OS(m.Dir), fsGate), "/", vfs.BindAfter)
	}

	corpus := godoc.NewCorpus(moduleFS{fs})
	corpus.IndexDirectory = func(dir string) bool {
		return dir != "/pkg" && !strings.HasPrefix(dir, "/pkg/")
	}
	if err := corpus.Init(); err != nil {
		return nil, nil, err
	}

	// Initialize the version info before readTemplates, which saves
	// the map value in a method value.
	corpus.InitVersionInfo()

	pres := godoc.NewPresentation(corpus)
	pres.DeclLinks = true

	if err := readTemplates(fs, pres); err != nil {
		return nil, nil, err
	}

	return pres, mods, nil
}

func Handler(goModFile string, static bool) (*http.ServeMux, error) {
	pres, mods, err := newPresentation(goModFile)
	if err != nil {
		return nil, err
	}
	return handlers(pres, mods, static), nil
}

func handlers(pres http.Handler, mods []mod, static bool) *http.ServeMux {
	modPaths := make(map[string]struct{}, len(mods))
	for _, mod := range mods {
		modPaths[mod.Path] = struct{}{}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			http.Redirect(w, req, "/pkg/", http.StatusFound)
			return
		}
		lrw := &responseWriter{
			ResponseWriter: w,
		}
		lrw.Before(func(beforeW *responseWriter) bool {
			if beforeW.Status() == 404 {
				if static {
					w.WriteHeader(http.StatusNotFound)
					return false
				}
				req.URL.Scheme = "https"
				req.URL.Host = "pkg.go.dev"
				http.Redirect(w, req, req.URL.String(), http.StatusPermanentRedirect)
				return false
			}
			return true
		})
		pres.ServeHTTP(lrw, req)
	})

	return mux
}

func readTemplates(fs vfs.Opener, p *godoc.Presentation) error {
	var err error
	p.CallGraphHTML, err = readTemplate(fs, p, "callgraph.html")
	if err != nil {
		return err
	}
	p.DirlistHTML, err = readTemplate(fs, p, "dirlist.html")
	if err != nil {
		return err
	}
	p.ErrorHTML, err = readTemplate(fs, p, "error.html")
	if err != nil {
		return err
	}
	p.ExampleHTML, err = readTemplate(fs, p, "example.html")
	if err != nil {
		return err
	}
	p.GodocHTML, err = readTemplate(fs, p, "godoc.html")
	if err != nil {
		return err
	}
	p.ImplementsHTML, err = readTemplate(fs, p, "implements.html")
	if err != nil {
		return err
	}
	p.MethodSetHTML, err = readTemplate(fs, p, "methodset.html")
	if err != nil {
		return err
	}
	p.PackageHTML, err = readTemplate(fs, p, "package.html")
	if err != nil {
		return err
	}
	p.PackageRootHTML, err = readTemplate(fs, p, "packageroot.html")
	if err != nil {
		return err
	}
	p.SearchHTML, err = readTemplate(fs, p, "search.html")
	if err != nil {
		return err
	}
	p.SearchDocHTML, err = readTemplate(fs, p, "searchdoc.html")
	if err != nil {
		return err
	}
	p.SearchCodeHTML, err = readTemplate(fs, p, "searchcode.html")
	if err != nil {
		return err
	}
	p.SearchTxtHTML, err = readTemplate(fs, p, "searchtxt.html")
	if err != nil {
		return err
	}
	return nil
}

func readTemplate(fs vfs.Opener, p *godoc.Presentation, name string) (*template.Template, error) {
	path := "lib/godoc/" + name

	// use underlying file system fs to read the template file
	// (cannot use template ParseFile functions directly)
	data, err := vfs.ReadFile(fs, path)
	if err != nil {
		return nil, err
	}
	// be explicit with errors (for app engine use)
	t, err := template.New(name).Funcs(p.FuncMap()).Parse(string(data))
	if err != nil {
		return nil, err
	}
	return t, nil
}

type mod struct {
	Main     bool
	Indirect bool
	Path     string // Module path.
	Dir      string // Directory holding files for this module, if any.
}

// buildList determines the build list in the current directory
// by invoking the go command. It should only be used in module mode,
// when vendor mode isn't on.
//
// See https://golang.org/cmd/go/#hdr-The_main_module_and_the_build_list.
func buildList(goModPath string) ([]mod, error) {
	if goModPath == os.DevNull {
		// Empty build list.
		return nil, nil
	}

	cmd := exec.Command("go", "list", "-m", "-json", "all")
	cmd.Dir = goModPath
	out, err := cmd.Output()
	if ee := (*exec.ExitError)(nil); xerrors.As(err, &ee) {
		return nil, fmt.Errorf("go command exited unsuccessfully: %v\n%s", ee.ProcessState.String(), ee.Stderr)
	} else if err != nil {
		return nil, err
	}
	var mods []mod
	for dec := json.NewDecoder(bytes.NewReader(out)); ; {
		var m mod
		err := dec.Decode(&m)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if m.Main || m.Indirect {
			continue
		}
		mods = append(mods, m)
	}
	return mods, nil
}

// moduleFS is a vfs.FileSystem wrapper used when godoc is running
// in module mode. It's needed so that packages inside modules are
// considered to be third party.
//
// It overrides the RootType method of the underlying filesystem
// and implements it using a heuristic based on the import path.
// If the first element of the import path does not contain a dot,
// that package is considered to be inside GOROOT. If it contains
// a dot, then that package is considered to be third party.
//
// TODO(dmitshur): The RootType abstraction works well when GOPATH
// workspaces are bound at their roots, but scales poorly in the
// general case. It should be replaced by a more direct solution
// for determining whether a package is third party or not.
//
type moduleFS struct{ vfs.FileSystem }

func (moduleFS) RootType(path string) vfs.RootType {
	if !strings.HasPrefix(path, "/src/") {
		return ""
	}
	domain := path[len("/src/"):]
	if i := strings.Index(domain, "/"); i >= 0 {
		domain = domain[:i]
	}
	if !strings.Contains(domain, ".") {
		// No dot in the first element of import path
		// suggests this is a package in GOROOT.
		return vfs.RootTypeGoRoot
	}
	// A dot in the first element of import path
	// suggests this is a third party package.
	return vfs.RootTypeGoPath
}
func (fs moduleFS) String() string { return "module(" + fs.FileSystem.String() + ")" }

// adapted from https://github.com/urfave/negroni/blob/7915ed3d6bdb7d071a72f9f1e15ace4c78e97636/response_writer.go
type responseWriter struct {
	http.ResponseWriter
	status      int
	size        int
	beforeFuncs []beforeFunc
	frozen      bool
}

type beforeFunc func(*responseWriter) bool

func (rw *responseWriter) WriteHeader(s int) {
	rw.status = s
	rw.callBefore()
	if rw.frozen {
		return
	}
	rw.ResponseWriter.WriteHeader(s)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.frozen {
		return 0, nil
	}
	if !rw.Written() {
		// The status will be StatusOK if WriteHeader has not been called yet
		rw.WriteHeader(http.StatusOK)
	}
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) Size() int {
	return rw.size
}

func (rw *responseWriter) Written() bool {
	return rw.status != 0
}

func (rw *responseWriter) Before(before func(*responseWriter) bool) {
	rw.beforeFuncs = append(rw.beforeFuncs, before)
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the responseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (rw *responseWriter) callBefore() {
	for i := len(rw.beforeFuncs) - 1; i >= 0; i-- {
		cont := rw.beforeFuncs[i](rw)
		if !cont {
			rw.frozen = true
		}
	}
}

func (rw *responseWriter) Flush() {
	if rw.frozen {
		return
	}
	flusher, ok := rw.ResponseWriter.(http.Flusher)
	if ok {
		if !rw.Written() {
			// The status will be StatusOK if WriteHeader has not been called yet
			rw.WriteHeader(http.StatusOK)
		}
		flusher.Flush()
	}
}
