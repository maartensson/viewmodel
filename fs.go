package viewmodel

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type FS struct {
	uuid  uuid.UUID
	fs    fs.FS
	files []string
}

func allFSs(vm VM) []fs.FS {
	if vm != nil {
		return append(allFSs(vm.Data()), vm.FS())
	}
	return []fs.FS{}
}

func htmls(fsys fs.FS) ([]string, error) {
	var htmlFiles []string

	if err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".html") {
			htmlFiles = append(htmlFiles, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return htmlFiles, nil
}

func newFS(fsys fs.FS) (*FS, error) {
	files, err := htmls(fsys)
	if err != nil {
		return nil, err
	}
	return &FS{
		uuid:  uuid.New(),
		files: files,
		fs:    fsys,
	}, nil
}

func traverse(vm VM) []*FS {
	if vm != nil {
		fsys, err := newFS(vm.FS())
		if err != nil {
			panic(err)
		}
		return append(traverse(vm.Data()), fsys)
	}
	return []*FS{}
}

type mfs struct{ fss map[uuid.UUID]fs.FS }

func Merge(fss []fs.FS) (fs.FS, []string) {
	var files []string
	fsys := make(map[uuid.UUID]fs.FS)
	for _, e := range fss {
		uuid := uuid.New()
		fsys[uuid] = e

		htmls, err := htmls(e)
		if err != nil {
			panic(err)
		}
		for _, e := range htmls {
			files = append(files, filepath.Join(uuid.String(), e))
		}

	}
	fmt.Println(files)
	return mfs{fss: fsys}, files
}

func (mfs mfs) Open(name string) (fs.File, error) {
	dirs := strings.Split(filepath.Clean(name), string(filepath.Separator))
	if len(dirs) == 0 {
		return nil, os.ErrNotExist
	}

	uuid, err := uuid.Parse(dirs[0])
	if err != nil {
		return nil, os.ErrNotExist
	}

	fs, ok := mfs.fss[uuid]
	if !ok {
		return nil, os.ErrNotExist
	}

	f, err := fs.Open(name)
	if err != nil {
		return nil, os.ErrNotExist
	}

	return f, nil
}
