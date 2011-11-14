package main

import (
	"log"
	"path/filepath"
	"os"
	"strings"
	"strconv"
	"github.com/hanwen/go-fuse/fuse"
)

type FuseVfs struct {
	fuse.DefaultFileSystem

	databasePath string
	mountPath    string
	state        *fuse.MountState
}

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

func (this FuseVfs) Unmount() {
	this.state.Unmount()
}

func (this FuseVfs) Loop() {
	this.state.Loop()
}

func (this FuseVfs) GetAttr(name string, context *fuse.Context) (*os.FileInfo, fuse.Status) {
	switch name {
	case "":
		fallthrough
	case "tags":
		return &os.FileInfo{Mode: fuse.S_IFDIR | 0755}, fuse.OK
	}

	path := this.splitPath(name)
	switch path[0] {
	case "tags":
		return this.getTaggedEntryAttr(path[1:])
	}

	return nil, fuse.ENOENT
}

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

func (this FuseVfs) Open(name string, flags uint32, context *fuse.Context) (fuse.File, fuse.Status) {
	//TODO
	//if flags & fuse.O_ANYWRITE != 0 { return nil, fuse.EPERM }

	return fuse.NewDataFile([]byte("tmsu (c) 2011 Paul Ruane\n")), fuse.OK
}

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
		return 0, error
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
		channel <- fuse.DirEntry{Name: tag.Name, Mode: fuse.S_IFREG}
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
		return &os.FileInfo{Mode: fuse.S_IFDIR | 0755}, fuse.OK
	}

	return &os.FileInfo{Mode: fuse.S_IFLNK | 0755, Size: int64(10)}, fuse.OK
}

func (this FuseVfs) openTaggedEntryDir(path []string) (chan fuse.DirEntry, fuse.Status) {
	db, error := OpenDatabase(this.databasePath)
	if error != nil {
		log.Fatalf("Could not open database: %v", error)
	}
	defer db.Close()

	//TODO assumption that all path dirs are tags
	tags := path

	furtherTags, error := db.TagsForTags(tags)
	if error != nil {
		log.Fatalf("Could not retrieve tags for tags: %v", error)
	}

	files, error := db.FilesWithTags(path)
	if error != nil {
		log.Fatalf("Could not retrieve tagged files: %v", error)
	}

	channel := make(chan fuse.DirEntry, len(*files)+len(*furtherTags))
	defer close(channel)

	for _, tag := range *furtherTags {
		channel <- fuse.DirEntry{Name: tag.Name, Mode: fuse.S_IFDIR | 0755}
	}

	for _, file := range *files {
		extension := filepath.Ext(file.Path)
		fileName := filepath.Base(file.Path)
		fileName = fileName[0 : len(fileName)-len(extension)]

		channel <- fuse.DirEntry{Name: fileName + "." + strconv.Uitoa(file.Id) + extension, Mode: fuse.S_IFLNK}
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
