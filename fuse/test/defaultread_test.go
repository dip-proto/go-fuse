// Copyright 2016 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"io/ioutil"
	"testing"

	"github.com/dip-proto/go-fuse/v2/fuse"
	"github.com/dip-proto/go-fuse/v2/fuse/nodefs"
	"github.com/dip-proto/go-fuse/v2/fuse/pathfs"
	"github.com/dip-proto/go-fuse/v2/internal/testutil"
)

type DefaultReadFS struct {
	pathfs.FileSystem
	size  uint64
	exist bool
}

func (fs *DefaultReadFS) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	if name == "" {
		return &fuse.Attr{Mode: fuse.S_IFDIR | 0755}, fuse.OK
	}
	if name == "file" {
		return &fuse.Attr{Mode: fuse.S_IFREG | 0644, Size: fs.size}, fuse.OK
	}
	return nil, fuse.ENOENT
}

func (fs *DefaultReadFS) Open(name string, f uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	return nodefs.NewDefaultFile(), fuse.OK
}

func defaultReadTest(t *testing.T) (root string, cleanup func()) {
	fs := &DefaultReadFS{
		FileSystem: pathfs.NewDefaultFileSystem(),
		size:       22,
	}

	var err error
	dir := t.TempDir()
	pathfs := pathfs.NewPathNodeFs(fs, nil)
	opts := nodefs.NewOptions()
	opts.Debug = testutil.VerboseTest()

	state, _, err := nodefs.MountRoot(dir, pathfs.Root(), opts)
	if err != nil {
		t.Fatalf("MountNodeFileSystem failed: %v", err)
	}
	go state.Serve()
	if err := state.WaitMount(); err != nil {
		t.Fatal("WaitMount", err)
	}
	return dir, func() {
		state.Unmount()
	}
}

func TestDefaultRead(t *testing.T) {
	root, clean := defaultReadTest(t)
	defer clean()

	_, err := ioutil.ReadFile(root + "/file")
	if err == nil {
		t.Fatal("should have failed read.")
	}
}
