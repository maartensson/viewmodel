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
}

func New[T VM](fs fs.FS, title string, data T) *rootModel {
	return &rootModel{
		Title: title,
		data:  data,
		fs:    fs,
	}
}

func (vm *rootModel) Execute(w http.ResponseWriter) {
	templ, err := template.
		New("index").
		Funcs(template.FuncMap{
			"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		}).
		ParseFS(mergefs.Merge(allFSs(vm)), allPaths(vm)...)
	if err != nil {
		panic(err)
	}

	if err := templ.Execute(w, &vm); err != nil {
		panic(err)
	}
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
	Title string
	data  VM
	fs    fs.FS
}

func (vm *rootModel) Data() VM        { return vm.data }
func (vm *rootModel) Templ() []string { return []string{"index.html"} }
func (vm *rootModel) FS() fs.FS       { return vm.fs }

type baseModel struct {
	fs    fs.FS
	paths []string
}

// If the viewmodel has values it is not basic
func Basic(fs fs.FS, paths ...string) *baseModel { return &baseModel{paths: paths, fs: fs} }
func (vm *baseModel) Templ() []string            { return vm.paths }
func (vm *baseModel) Data() VM                   { return nil }
func (vm *baseModel) FS() fs.FS                  { return vm.fs }
