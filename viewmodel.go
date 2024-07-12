package viewmodel

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/mamaart/mergefs"
)

type VM interface {
	FS() fs.FS
	Templ() []string
	Data() VM
	Title() string
}

func New[T VM](fs fs.FS, title string, data T) Root {
	return Raw("index", fs, &rootModel{
		title: title,
		data:  data,
		fs:    fs,
	})
}

func allPaths(vm VM) []string {
	if vm != nil {
		return append(vm.Templ(), allPaths(vm.Data())...)
	}
	return []string{}
}

func allFSs(vm VM) []fs.FS {
	if vm != nil {
		return append(allFSs(vm.Data()), vm.FS())
	}
	return []fs.FS{}
}

type rootModel struct {
	title string
	data  VM
	fs    fs.FS
}

func (vm *rootModel) Data() VM        { return vm.data }
func (vm *rootModel) Templ() []string { return []string{"index.html"} }
func (vm *rootModel) FS() fs.FS       { return vm.fs }
func (vm *rootModel) Title() string   { return vm.title }

type baseModel struct {
	fs    fs.FS
	title string
	paths []string
}

// If the viewmodel has values it is not basic
func Basic(fs fs.FS, title string, paths ...string) *baseModel {
	return &baseModel{paths: paths, title: title, fs: fs}
}
func (vm *baseModel) Templ() []string { return vm.paths }
func (vm *baseModel) Data() VM        { return nil }
func (vm *baseModel) FS() fs.FS       { return vm.fs }
func (vm *baseModel) Title() string   { return vm.title }

type raw struct {
	name  string
	fs    fs.FS
	inner VM
}

type Root interface {
	Execute(w http.ResponseWriter)
}

func Raw[T VM](name string, fs fs.FS, data T) Root {
	return &raw{inner: data, fs: fs, name: name}
}

func (raw *raw) Execute(w http.ResponseWriter) {
	templ, err := template.
		New(raw.name).
		Funcs(template.FuncMap{
			"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		}).
		ParseFS(mergefs.Merge(allFSs(raw.inner)), allPaths(raw.inner)...)
	if err != nil {
		panic(err)
	}

	if err := templ.Execute(w, &raw.inner); err != nil {
		panic(err)
	}
}
