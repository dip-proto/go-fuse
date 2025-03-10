// Copyright 2016 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux
// +build linux

package test

import (
	"path/filepath"
	"syscall"
	"testing"

	"github.com/dip-proto/go-fuse/v2/fuse"
	"github.com/dip-proto/go-fuse/v2/fuse/nodefs"
	"github.com/dip-proto/go-fuse/v2/internal/testutil"
)

// this file is linux-only, since it uses syscall.Getxattr.

type xattrNode struct {
	nodefs.Node
}

func (n *xattrNode) OnMount(fsConn *nodefs.FileSystemConnector) {
	n.Inode().NewChild("child", false, &xattrChildNode{nodefs.NewDefaultNode()})
}

type xattrChildNode struct {
	nodefs.Node
}

func (n *xattrChildNode) GetXAttr(attr string, context *fuse.Context) ([]byte, fuse.Status) {
	if attr == "attr" {
		return []byte("value"), fuse.OK
	}
	return []byte(""), fuse.OK
}

func TestDefaultXAttr(t *testing.T) {
	dir := t.TempDir()

	root := &xattrNode{
		Node: nodefs.NewDefaultNode(),
	}

	opts := nodefs.NewOptions()
	opts.Debug = testutil.VerboseTest()
	s, _, err := nodefs.MountRoot(dir, root, opts)
	if err != nil {
		t.Fatalf("MountRoot: %v", err)
	}
	go s.Serve()
	if err := s.WaitMount(); err != nil {
		t.Fatal("WaitMount", err)
	}

	defer s.Unmount()

	var data [1024]byte
	sz, err := syscall.Getxattr(filepath.Join(dir, "child"), "attr", data[:])
	if err != nil {
		t.Fatalf("Getxattr: %v", err)
	} else if got, want := string(data[:sz]), "value"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestEmptyXAttr(t *testing.T) {
	dir := t.TempDir()

	root := &xattrNode{
		Node: nodefs.NewDefaultNode(),
	}

	opts := nodefs.NewOptions()
	opts.Debug = testutil.VerboseTest()
	s, _, err := nodefs.MountRoot(dir, root, opts)
	if err != nil {
		t.Fatalf("MountRoot: %v", err)
	}
	go s.Serve()
	if err := s.WaitMount(); err != nil {
		t.Fatal("WaitMount", err)
	}

	defer s.Unmount()

	var data [1024]byte
	sz, err := syscall.Getxattr(filepath.Join(dir, "child"), "attr2", data[:])
	if err != nil {
		t.Fatalf("Getxattr: %v", err)
	} else if got, want := string(data[:sz]), ""; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
