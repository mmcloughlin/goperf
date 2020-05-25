package dashboard

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"sync"

	"github.com/mmcloughlin/goperf/pkg/fs"
)

// Templates manages a set of templates with a base layout.
type Templates struct {
	fs         fs.Readable
	layoutsdir string
	cache      bool

	layouts   *template.Template
	templates sync.Map
}

// NewTemplates initializes a template collection.
func NewTemplates(r fs.Readable) *Templates {
	return &Templates{
		fs:         r,
		layoutsdir: "layout",
		cache:      true,
		layouts:    template.New("layouts"),
	}
}

// SetCacheEnabled configures whether templates are cached.
func (t *Templates) SetCacheEnabled(enabled bool) {
	t.cache = enabled
}

// Func declares a template function.
func (t *Templates) Func(name string, f interface{}) {
	t.layouts.Funcs(map[string]interface{}{
		name: f,
	})
}

func (t *Templates) Init(ctx context.Context) error {
	// Load layouts.
	files, err := t.fs.List(ctx, t.layoutsdir)
	if err != nil {
		return err
	}

	for _, file := range files {
		b, err := fs.ReadFile(ctx, t.fs, file.Path)
		if err != nil {
			return fmt.Errorf("read file %q: %w", file.Path, err)
		}

		if _, err := t.layouts.Parse(string(b)); err != nil {
			return err
		}
	}

	return nil
}

// Template parses and returns a template file. The template will be returned from a cache if it has been loaded before.
func (t *Templates) Template(ctx context.Context, path string) (*template.Template, error) {
	if tmpl, ok := t.templates.Load(path); ok {
		return tmpl.(*template.Template), nil
	}

	b, err := fs.ReadFile(ctx, t.fs, path)
	if err != nil {
		return nil, fmt.Errorf("load template %q: %w", path, err)
	}

	tmpl, err := t.layouts.Clone()
	if err != nil {
		return nil, err
	}

	if _, err := tmpl.Parse(string(b)); err != nil {
		return nil, err
	}

	if t.cache {
		t.templates.Store(path, tmpl)
	}

	return tmpl, nil
}

// ExecuteTemplate loads and renders the given template to w.
func (t *Templates) ExecuteTemplate(ctx context.Context, w io.Writer, path, name string, data interface{}) error {
	tmpl, err := t.Template(ctx, path)
	if err != nil {
		return err
	}

	return tmpl.ExecuteTemplate(w, name, data)
}
