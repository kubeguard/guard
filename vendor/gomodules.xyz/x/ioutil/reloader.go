package ioutil

import (
	"bytes"
	"embed"
	iofs "io/fs"
	"os"
)

const (
	TriggerFile = "trigger"
)

type Reloader struct {
	dir string
	fs  embed.FS

	trigger []byte
	loaded  bool
	loadFn  func(fsys iofs.FS)
}

func NewReloader(dir string, fs embed.FS, loadFn func(fsys iofs.FS)) *Reloader {
	return &Reloader{
		dir:     dir,
		fs:      fs,
		trigger: nil,
		loaded:  false,
		loadFn:  loadFn,
	}
}

func (r *Reloader) FS() iofs.FS {
	if fi, err := os.Stat(r.dir); os.IsNotExist(err) || !fi.IsDir() {
		return r.fs
	}
	return os.DirFS(r.dir)
}

func (r *Reloader) needsReload(fsys iofs.FS) bool {
	if data, err := iofs.ReadFile(fsys, TriggerFile); err == nil {
		yes := bytes.Compare(r.trigger, data) != 0
		r.trigger = data
		return yes || !r.loaded // ensure loads at least first time
	}
	return !r.loaded
}

func (r *Reloader) ReloadIfTriggered() {
	fsys := r.FS()
	if r.needsReload(fsys) {
		r.loadFn(fsys)
		r.loaded = true
	}
}
