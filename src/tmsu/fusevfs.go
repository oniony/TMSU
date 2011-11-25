// Copyright 2011 Paul Ruane. All rights reserved.

package main

import (
	"log"
	"path/filepath"
	"os"
	"strings"
	"strconv"
	"time"
	"github.com/hanwen/go-fuse/fuse"
)

type FuseVfs struct {
	fuse.DefaultFileSystem

	databasePath string
	mountPath    string
	state        *fuse.MountState
}

// Mount the VFS.
func MountVfs(databasePath string, mountPath string) (*FuseVfs, error) {
	fuseVfs := FuseVfs{}

	state, _, error := fuse.MountPathFileSystem(mountPath, &fuseVfs, nil)
	if error != nil {
		return nil, error
	}

	fuseVfs.databasePath = databasePath
	fuseVfs.mountPath = mountPath
	fuseVfs.state = state

	return &fuseVfs, nil
}

// Unmount the VFS.
func (this FuseVfs) Unmount() {
	this.state.Unmount()
}

func (this FuseVfs) Loop() {
	this.state.Loop()
}

// Get entry attributes.
func (this FuseVfs) GetAttr(name string, context *fuse.Context) (*os.FileInfo, fuse.Status) {
	switch name {
	case "":
		fallthrough
	case "tags":
        now := time.Nanoseconds()
		return &os.FileInfo{Mode: fuse.S_IFDIR | 0755, Atime_ns: now, Mtime_ns: now, Ctime_ns: now}, fuse.OK
	}

	path := this.splitPath(name)
	switch path[0] {
	case "tags":
		return this.getTaggedEntryAttr(path[1:])
	}

	return nil, fuse.ENOENT
}

// Unlink entry.
func (this FuseVfs) Unlink(name string, context *fuse.Context) fuse.Status {
    fileId, error := this.parseFileId(name)
    if error != nil {
        log.Fatalf("Could not unlink: %v", error)
    }

    if fileId == 0 {
        // cannot unlink tag directories
        return fuse.EPERM
    }

    path := this.splitPath(name)
    tagNames := path[1:len(path) - 1]

    db, error := OpenDatabase(databasePath())
    if error != nil {
        log.Fatal(error)
    }

    for _, tagName := range tagNames {
        tag, error := db.TagByName(tagName)
        if error != nil {
            log.Fatal(error)
        }
        if tag == nil {
            log.Fatalf("Could not retrieve tag '%v'.", tagName)
        }

        error = db.RemoveFileTag(fileId, tag.Id)
        if error != nil {
            log.Fatal(error)
        }
    }

    return fuse.OK
}

// Enumerate directory.
func (this FuseVfs) OpenDir(name string, context *fuse.Context) (chan fuse.DirEntry, fuse.Status) {
	switch name {
	case "":
		return this.topDirectories()
	case "tags":
		return this.tagDirectories()
	}

	path := this.splitPath(name)
	switch path[0] {
	case "tags":
		return this.openTaggedEntryDir(path[1:])
	}

	return nil, fuse.ENOENT
}

// Read symbolic link target.
func (this FuseVfs) Readlink(name string, context *fuse.Context) (string, fuse.Status) {
	path := this.splitPath(name)
	switch path[0] {
	case "tags":
		return this.readTaggedEntryLink(path[1:])
	}

	return "", fuse.ENOENT
}

// implementation

func (this FuseVfs) splitPath(path string) []string {
	return strings.Split(path, string(filepath.Separator))
}

func (this FuseVfs) parseFileId(name string) (uint, error) {
	parts := strings.Split(name, ".")
	count := len(parts)

	if count == 1 {
		return 0, nil
	}

	id, error := strconv.Atoui(parts[count-2])
	if error != nil {
		id, error = strconv.Atoui(parts[count-1])
	}
	if error != nil {
		return 0, nil
	}

	return id, nil
}

func (this FuseVfs) topDirectories() (chan fuse.DirEntry, fuse.Status) {
	channel := make(chan fuse.DirEntry, 2)
	channel <- fuse.DirEntry{Name: "tags", Mode: fuse.S_IFDIR}
	close(channel)

	return channel, fuse.OK
}

func (this FuseVfs) tagDirectories() (chan fuse.DirEntry, fuse.Status) {
	db, error := OpenDatabase(this.databasePath)
	if error != nil {
		log.Fatal("Could not open database: %v", error)
	}
	defer db.Close()

	tags, error := db.Tags()
	if error != nil {
		log.Fatal("Could not retrieve tags: %v", error)
	}

	channel := make(chan fuse.DirEntry, len(*tags))
	for _, tag := range *tags {
		channel <- fuse.DirEntry{Name: tag.Name, Mode: fuse.S_IFDIR}
	}
	close(channel)

	return channel, fuse.OK
}

func (this FuseVfs) getTaggedEntryAttr(path []string) (*os.FileInfo, fuse.Status) {
	pathLength := len(path)
	name := path[pathLength-1]

	fileId, error := this.parseFileId(name)
	if error != nil {
		return nil, fuse.ENOENT
	}

	if fileId == 0 {
	    // if no file ID then it is a tag directory
        now := time.Nanoseconds()
		return &os.FileInfo{Mode: fuse.S_IFDIR | 0755, Atime_ns: now, Mtime_ns: now, Ctime_ns: now}, fuse.OK
	}

	db, error := OpenDatabase(this.databasePath)
	if error != nil {
		log.Fatalf("Could not open database: %v", error)
	}
	defer db.Close()

    file, error := db.File(fileId)
    if error != nil {
        log.Fatalf("Could not retrieve file #%v: %v", fileId, error)
    }
    if file == nil {
        return &os.FileInfo{Mode: fuse.S_IFREG}, fuse.ENOENT
    }

    fileInfo, error := os.Stat(file.Path)
    var atime, mtime, ctime, size int64
    if error == nil {
        atime = fileInfo.Atime_ns
        mtime = fileInfo.Mtime_ns
        ctime = fileInfo.Ctime_ns
        size = fileInfo.Size
    } else {
        now := time.Nanoseconds()
        atime = now
        mtime = now
        ctime = now
        size = 0
    }

	return &os.FileInfo{Mode: fuse.S_IFLNK | 0755, Atime_ns: atime, Mtime_ns: mtime, Ctime_ns: ctime, Size: size}, fuse.OK
}

func (this FuseVfs) openTaggedEntryDir(path []string) (chan fuse.DirEntry, fuse.Status) {
	db, error := OpenDatabase(this.databasePath)
	if error != nil {
		log.Fatalf("Could not open database: %v", error)
	}
	defer db.Close()

	tags := path

	furtherTags, error := db.TagsForTags(tags)
	if error != nil {
		log.Fatalf("Could not retrieve tags for tags: %v", error)
	}

	files, error := db.FilesWithTags(tags)
	if error != nil {
		log.Fatalf("Could not retrieve tagged files: %v", error)
	}

	channel := make(chan fuse.DirEntry, len(*files)+len(*furtherTags))
	defer close(channel)

	for _, tag := range *furtherTags {
		channel <- fuse.DirEntry{Name: tag.Name, Mode: fuse.S_IFDIR | 0755}
	}

	for _, file := range *files {
        linkName := this.getLinkName(file)
		channel <- fuse.DirEntry{Name: linkName, Mode: fuse.S_IFLNK}
	}

	return channel, fuse.OK
}

func (this FuseVfs) readTaggedEntryLink(path []string) (string, fuse.Status) {
	name := path[len(path)-1]

	db, error := OpenDatabase(this.databasePath)
	if error != nil {
		log.Fatalf("Could not open database: %v", error)
	}
	defer db.Close()

	fileId, error := this.parseFileId(name)
	if error != nil {
		log.Fatalf("Could not parse file identifier: %v", error)
	}
	if fileId == 0 {
		return "", fuse.ENOENT
	}

	file, error := db.File(fileId)
	if error != nil {
		log.Fatalf("Could not find file %v in database.", fileId)
	}

	return file.Path, fuse.OK
}

func (this FuseVfs) getLinkName(file File) string {
    extension := filepath.Ext(file.Path)
    fileName := filepath.Base(file.Path)
    linkName := fileName[0 : len(fileName) - len(extension)]
    suffix := "." + strconv.Uitoa(file.Id) + extension

    if len(linkName) + len(suffix) > 255 {
        linkName = linkName[0 : 255 - len(suffix)]
    }

    return linkName + suffix
}
