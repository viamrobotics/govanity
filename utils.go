package govanity

import (
	"embed"
	"path/filepath"
	"runtime"
)

//EmbeddedTemplates are the main templates for the app
//go:embed templates/*.html
var EmbeddedTemplates embed.FS

func ResolveFile(fn string) string {
	_, thisFilePath, _, _ := runtime.Caller(0)
	thisDirPath, err := filepath.Abs(filepath.Dir(thisFilePath))
	if err != nil {
		panic(err)
	}
	return filepath.Join(thisDirPath, fn)
}
