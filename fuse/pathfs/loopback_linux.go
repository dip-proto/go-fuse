// Copyright 2016 the Go-FUSE Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathfs

import (
	"fmt"
	"syscall"
	"time"

	"github.com/dip-proto/go-fuse/v2/fuse"
)

func (fs *loopbackFileSystem) ListXAttr(name string, context *fuse.Context) ([]string, fuse.Status) {
	attrs, err := listXAttr(fs.GetPath(name))
	return attrs, fuse.ToStatus(err)
}

func (fs *loopbackFileSystem) RemoveXAttr(name string, attr string, context *fuse.Context) fuse.Status {
	err := syscall.Removexattr(fs.GetPath(name), attr)
	return fuse.ToStatus(err)
}

func (fs *loopbackFileSystem) String() string {
	return fmt.Sprintf("LoopbackFs(%s)", fs.Root)
}

func (fs *loopbackFileSystem) GetXAttr(name string, attr string, context *fuse.Context) ([]byte, fuse.Status) {

	bufsz := 1024
	for {
		data := make([]byte, bufsz)
		sz, err := syscall.Getxattr(fs.GetPath(name), attr, data)
		if err == nil {
			return data[:sz], fuse.OK
		}
		if err == syscall.ERANGE {
			bufsz = sz
			continue
		}

		return nil, fuse.ToStatus(err)
	}

}

func (fs *loopbackFileSystem) SetXAttr(name string, attr string, data []byte, flags int, context *fuse.Context) fuse.Status {
	err := syscall.Setxattr(fs.GetPath(name), attr, data, flags)
	return fuse.ToStatus(err)
}

// Utimens - path based version of loopbackFile.Utimens()
func (fs *loopbackFileSystem) Utimens(path string, a *time.Time, m *time.Time, context *fuse.Context) (code fuse.Status) {
	var ts [2]syscall.Timespec
	ts[0] = fuse.UtimeToTimespec(a)
	ts[1] = fuse.UtimeToTimespec(m)
	err := sysUtimensat(0, fs.GetPath(path), &ts, _AT_SYMLINK_NOFOLLOW)
	return fuse.ToStatus(err)
}
