package main

import (
	"os"

	"github.com/boltdb/bolt"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

// Root is a Bolt transaction root bucket. It's special because it
// cannot contain keys, and doesn't really have a *bolt.Bucket.
type Root struct {
	fs *FS
}

var _ = fs.Node(&Root{})

func (r *Root) Attr() fuse.Attr {
	return fuse.Attr{Inode: 1, Mode: os.ModeDir | 0755}
}

var _ = fs.HandleReadDirer(&Root{})

func (r *Root) ReadDir(intr fs.Intr) ([]fuse.Dirent, fuse.Error) {
	var res []fuse.Dirent
	err := r.fs.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			res = append(res, fuse.Dirent{
				Type: fuse.DT_Dir,
				Name: string(name),
			})
			return nil
		})
	})
	return res, err
}

var _ = fs.NodeStringLookuper(&Root{})

func (r *Root) Lookup(name string, intr fs.Intr) (fs.Node, fuse.Error) {
	var n fs.Node
	err := r.fs.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(name))
		if b == nil {
			return fuse.ENOENT
		}
		n = &Dir{
			root: r,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return n, nil
}
