package zinc

import (
	"io/fs"
	"net/http"
	"os"
)

type StaticConfig struct {
	Browse bool
	Index  string
}

type StaticOption func(*StaticConfig)

func WithStaticBrowse(browse bool) StaticOption {
	return func(cfg *StaticConfig) {
		cfg.Browse = browse
	}
}

func WithStaticIndex(index string) StaticOption {
	return func(cfg *StaticConfig) {
		cfg.Index = index
	}
}

func (a *App) Static(prefix, root string, opts ...StaticOption) error {
	return a.StaticFS(prefix, os.DirFS(root), opts...)
}

func (a *App) StaticFS(prefix string, filesystem fs.FS, opts ...StaticOption) error {
	cfg := StaticConfig{Index: "index.html"}
	for _, opt := range opts {
		opt(&cfg)
	}
	_ = cfg
	a.Mount(prefix, http.FileServer(http.FS(filesystem)))
	return nil
}

func (a *App) File(path, file string) error {
	return a.Get(path, func(c *Context) error {
		return c.File(file)
	})
}

func (a *App) FileFS(path, file string, filesystem fs.FS) error {
	return a.Get(path, func(c *Context) error {
		return c.FileFS(file, filesystem)
	})
}
