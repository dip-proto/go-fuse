// Copyright 2016 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package nodefs

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/dip-proto/go-fuse/v2/fuse"
	"github.com/dip-proto/go-fuse/v2/internal/testutil"
)

const testTtl = 100 * time.Millisecond

func setupMemNodeTest(t *testing.T) (wd string, root Node, clean func()) {
	tmp, err := ioutil.TempDir("", "go-fuse-memnode_test")
	if err != nil {
		t.Fatalf("TempDir failed: %v", err)
	}
	back := tmp + "/backing"
	os.Mkdir(back, 0700)
	root = NewMemNodeFSRoot(back)
	mnt := tmp + "/mnt"
	os.Mkdir(mnt, 0700)

	connector := NewFileSystemConnector(root,
		&Options{
			EntryTimeout:        testTtl,
			AttrTimeout:         testTtl,
			NegativeTimeout:     0.0,
			Debug:               testutil.VerboseTest(),
			LookupKnownChildren: true,
		})
	state, err := fuse.NewServer(connector.RawFS(), mnt, &fuse.MountOptions{Debug: testutil.VerboseTest()})
	if err != nil {
		t.Fatal("NewServer", err)
	}

	// Unthreaded, but in background.
	go state.Serve()

	if err := state.WaitMount(); err != nil {
		t.Fatal("WaitMount", err)
	}
	return mnt, root, func() {
		state.Unmount()
		os.RemoveAll(tmp)
	}
}

func TestMemNodeFsWrite(t *testing.T) {
	wd, _, clean := setupMemNodeTest(t)
	defer clean()
	want := "hello"

	err := ioutil.WriteFile(wd+"/test", []byte(want), 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	content, err := ioutil.ReadFile(wd + "/test")
	if string(content) != want {
		t.Fatalf("content mismatch: got %q, want %q", content, want)
	}
}

func TestMemNodeFsBasic(t *testing.T) {
	wd, _, clean := setupMemNodeTest(t)
	defer clean()

	err := ioutil.WriteFile(wd+"/test", []byte{42}, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	fi, err := os.Lstat(wd + "/test")
	if err != nil {
		t.Fatalf("Lstat failed: %v", err)
	}
	if fi.Size() != 1 {
		t.Errorf("Size after write incorrect: got %d want 1", fi.Size())
	}

	entries, err := ioutil.ReadDir(wd)
	if len(entries) != 1 || entries[0].Name() != "test" {
		t.Fatalf("Readdir got %v, expected 1 file named 'test'", entries)
	}
}

func TestMemNodeSetattr(t *testing.T) {
	wd, _, clean := setupMemNodeTest(t)
	defer clean()

	f, err := os.OpenFile(wd+"/test", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer f.Close()

	err = f.Truncate(4096)
	if err != nil {
		t.Fatalf("Truncate failed: %v", err)
	}

	fi, err := f.Stat()
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if fi.Size() != 4096 {
		t.Errorf("Size should be 4096 after Truncate: %d", fi.Size())
	}

	if err := f.Chown(21, 42); err != nil {
		t.Errorf("Chown: %v", err)
	}
	if fi, err := f.Stat(); err != nil {
		t.Fatalf("Stat failed: %v", err)
	} else {
		attr := fuse.ToStatT(fi)
		if attr.Gid != 42 || attr.Uid != 21 {
			t.Errorf("got (%d, %d), want 42,21", attr.Uid, attr.Gid)
		}
	}
}
