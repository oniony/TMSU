package vfs

import (
          "github.com/hanwen/go-fuse/fuse"
          "os"
       )

func Mount(path string) (*Vfs, os.Error) {
    fuseVfs := FuseVfs{}
    state, _, error := fuse.MountPathFileSystem(path, &fuseVfs, nil)

    if error != nil {
        return nil, error
    }

    vfs := Vfs{
                  &fuseVfs,
                  state,
              }

    return &vfs, nil
}

type Vfs struct {
    fuseVfs *FuseVfs
    state *fuse.MountState
}

func (this *Vfs) Loop() {
    this.state.Loop()
}

type FuseVfs struct {
    fuse.DefaultFileSystem
}

func (this *FuseVfs) GetAttr(name string, context *fuse.Context) (*os.FileInfo, fuse.Status) {
    switch (name) {
        case "file.txt":
            return &os.FileInfo{
                                   Mode: fuse.S_IFREG | 0644,
                                   Size: int64(len(name)),
                               }, fuse.OK
        case "":
            return &os.FileInfo{
                                   Mode: fuse.S_IFDIR | 0755,
                               }, fuse.OK
    }

    return nil, fuse.ENOENT
}

func (this *FuseVfs) OpenDir(name string, context *fuse.Context) (chan fuse.DirEntry, fuse.Status) {
    if name != "" {
        return nil, fuse.ENOENT
    }

    channel := make(chan fuse.DirEntry, 1)
    channel <- fuse.DirEntry{
                                Name: "file.txt",
                                Mode: fuse.S_IFREG,
                            }
    close(channel)

    return channel, fuse.OK
}

func (this *FuseVfs) Open(name string, flags uint32, context *fuse.Context) (fuse.File, fuse.Status) {
    if name != "file.txt" {
            return nil, fuse.ENOENT
    }

    if flags & fuse.O_ANYWRITE != 0 {
        return nil, fuse.EPERM
    }

    return fuse.NewDataFile([]byte(name)), fuse.OK
}
