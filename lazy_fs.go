package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/djherbis/times"
)

type lazyFileMeta struct {
	entry            fs.DirEntry
	infoLoaded       bool
	linkResolved     bool
	linkTargetLoaded bool
	dirCountLoaded   bool
	timesLoaded      bool
	extLoaded        bool
}

func newFileEntry(path string, entry fs.DirEntry) *file {
	return &file{
		lazy:       lazyFileMeta{entry: entry, linkResolved: entry.Type()&os.ModeSymlink == 0},
		linkState:  notLink,
		path:       path,
		dirCount:   -1,
		dirSize:    -1,
		accessTime: time.Unix(0, 0),
		birthTime:  time.Unix(0, 0),
		changeTime: time.Unix(0, 0),
	}
}

func (f *file) isSymlink() bool {
	if f.lazy.entry != nil {
		return f.lazy.entry.Type()&os.ModeSymlink != 0
	}
	return f.FileInfo != nil && f.FileInfo.Mode()&os.ModeSymlink != 0
}

func (f *file) ensureLink() {
	if f.lazy.linkResolved {
		return
	}

	if !f.isSymlink() {
		f.lazy.linkResolved = true
		f.linkState = notLink
		return
	}

	if info, err := os.Stat(f.path); err == nil {
		f.FileInfo = info
		f.lazy.infoLoaded = true
		f.linkState = working
	} else {
		// Keep the entry lazy. A later Mode/Size/Sys call will fetch lstat
		// information only if it is actually needed.
		f.linkState = broken
		if !os.IsNotExist(err) {
			log.Printf("getting link target information: %s", err)
		}
	}
	f.lazy.linkResolved = true
}

func (f *file) ensureInfo() os.FileInfo {
	if f.isSymlink() && !f.lazy.linkResolved {
		f.ensureLink()
	}

	if f.FileInfo != nil {
		return f.FileInfo
	}
	if f.lazy.infoLoaded {
		return nil
	}
	if f.lazy.entry == nil {
		f.lazy.infoLoaded = true
		return nil
	}

	info, err := f.lazy.entry.Info()
	if err != nil {
		f.err = err
		f.lazy.infoLoaded = true
		return nil
	}

	f.FileInfo = info
	f.lazy.infoLoaded = true
	return info
}

func (f *file) Name() string {
	if f.lazy.entry != nil {
		return f.lazy.entry.Name()
	}
	if f.FileInfo != nil {
		return f.FileInfo.Name()
	}
	return filepath.Base(f.path)
}

func (f *file) Size() int64 {
	if info := f.ensureInfo(); info != nil {
		return info.Size()
	}
	return 0
}

func (f *file) Mode() os.FileMode {
	if info := f.ensureInfo(); info != nil {
		return info.Mode()
	}
	return 0
}

func (f *file) ModTime() time.Time {
	if info := f.ensureInfo(); info != nil {
		return info.ModTime()
	}
	return time.Unix(0, 0)
}

func (f *file) IsDir() bool {
	if f.lazy.entry != nil && !f.isSymlink() {
		if typ := f.lazy.entry.Type(); typ != 0 {
			return typ.IsDir()
		}
		if info := f.ensureInfo(); info != nil {
			return info.IsDir()
		}
		return false
	}

	f.ensureLink()
	if f.FileInfo != nil {
		return f.FileInfo.IsDir()
	}
	return false
}

func (f *file) Sys() any {
	if info := f.ensureInfo(); info != nil {
		return info.Sys()
	}
	return nil
}

func (f *file) modeType() os.FileMode {
	if f.lazy.entry != nil {
		if typ := f.lazy.entry.Type(); typ != 0 {
			return typ
		}
	}
	if info := f.ensureInfo(); info != nil {
		return info.Mode()
	}
	return 0
}

func (f *file) ensureTimes() {
	if f.lazy.timesLoaded {
		return
	}

	info := f.ensureInfo()
	if info == nil {
		f.lazy.timesLoaded = true
		return
	}

	ts := times.Get(info)
	f.accessTime = ts.AccessTime()

	f.birthTime = info.ModTime()
	if ts.HasBirthTime() {
		f.birthTime = ts.BirthTime()
	}

	f.changeTime = info.ModTime()
	if ts.HasChangeTime() {
		f.changeTime = ts.ChangeTime()
	}

	f.lazy.timesLoaded = true
}

func (f *file) ensureDirCount() int {
	if f.lazy.dirCountLoaded {
		return f.dirCount
	}

	if !getDirCounts(filepath.Dir(f.path)) || !f.IsDir() {
		return f.dirCount
	}

	d, err := os.Open(f.path)
	if err != nil {
		f.lazy.dirCountLoaded = true
		log.Printf("opening directory for count: %s", err)
		return f.dirCount
	}
	names, err := d.Readdirnames(10000)
	d.Close()

	if names == nil && err != nil {
		log.Printf("reading directory count: %s", err)
	} else {
		f.dirCount = len(names)
	}
	f.lazy.dirCountLoaded = true
	return f.dirCount
}

func (f *file) ensureLinkTarget() {
	if f.lazy.linkTargetLoaded {
		return
	}
	f.lazy.linkTargetLoaded = true

	if !f.isSymlink() {
		return
	}

	target, err := os.Readlink(f.path)
	if err != nil {
		log.Printf("reading link target: %s", err)
		return
	}
	f.linkTarget = target
}

func (f *file) extension() string {
	if f.lazy.extLoaded {
		return f.ext
	}

	if f.IsDir() {
		f.ext = ""
	} else {
		name := f.Name()
		if !(len(name) > 1 && name[0] == '.' && filepath.Ext(name) == "") {
			f.ext = filepath.Ext(name)
		}
	}
	f.lazy.extLoaded = true
	return f.ext
}
