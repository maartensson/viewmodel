package viewmodel

import (
	"html/template"
	"io/fs"
	"net/http"
)

type VM interface {
	FS() fs.FS
	Data() VM
}

func New[T VM](name string, fs fs.FS, data T) Root {
	return &raw{inner: data, fs: fs, name: name}
}

// If the viewmodel has NO values it is basic
type baseModel struct{ fs fs.FS }

func Basic(fs fs.FS) *baseModel { return &baseModel{fs: fs} }
func (vm *baseModel) Data() VM  { return nil }
func (vm *baseModel) FS() fs.FS { return vm.fs }

type raw struct {
	name  string
	fs    fs.FS
	inner VM
}

type Root interface{ Execute(w http.ResponseWriter) }

func (raw *raw) Execute(w http.ResponseWriter) {
	fs, paths := Merge(allFSs(raw.inner))
	templ, err := template.
		New(raw.name).
		Funcs(template.FuncMap{
			"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		}).
		ParseFS(fs, paths...)
	if err != nil {
		panic(err)
	}

	if err := templ.Execute(w, &raw.inner); err != nil {
		panic(err)
	}
}
