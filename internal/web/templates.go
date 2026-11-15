package web

import (
	"embed"
	"io/fs"
)

//go:embed templates/*.html
var templatesFS embed.FS

// Templates returns HTML templates for server-rendered pages embedded into binary, rooted at templates directory
func Templates() fs.FS {
	sub, err := fs.Sub(templatesFS, "templates")
	if err != nil {
		// fs.Sub fails only on invalid path, which is a constant here
		panic(err)
	}
	return sub
}
