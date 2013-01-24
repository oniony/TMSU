/*
Copyright 2011-2013 Paul Ruane.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package vfs

import (
	"fmt"
	"github.com/hanwen/go-fuse/fuse"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"tmsu/log"
	"tmsu/storage"
	"tmsu/storage/database"
)

type FuseVfs struct {
	fuse.DefaultFileSystem

	databasePath string
	mountPath    string
	state        *fuse.MountState
}

func MountVfs(databasePath string, mountPath string) (*FuseVfs, error) {
	fuseVfs := FuseVfs{}
	pathNodeFs := fuse.NewPathNodeFs(&fuseVfs, nil)
	state, _, err := fuse.MountNodeFileSystem(mountPath, pathNodeFs, nil)
	if err != nil {
		return nil, fmt.Errorf("could not mount virtual filesystem at '%v': %v", mountPath, err)
	}

	fuseVfs.databasePath = databasePath
	fuseVfs.mountPath = mountPath
	fuseVfs.state = state

	return &fuseVfs, nil
}

func (vfs FuseVfs) Unmount() {
	vfs.state.Unmount()
}

func (vfs FuseVfs) Loop() {
	vfs.state.Loop()
}

func (vfs FuseVfs) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	switch name {
	case "":
		fallthrough
	case "tags":
		return vfs.getTagsAttr()
	}

	path := vfs.splitPath(name)

	switch path[0] {
	case "tags":
		return vfs.getTaggedEntryAttr(path[1:])
	}

	return nil, fuse.ENOENT
}

func (vfs FuseVfs) Unlink(name string, context *fuse.Context) fuse.Status {
	fileId, err := vfs.parseFileId(name)
	if err != nil {
		log.Fatal("Could not unlink: ", err)
	}

	if fileId == 0 {
		// cannot unlink tag directories
		return fuse.EPERM
	}

	path := vfs.splitPath(name)
	tagNames := path[1 : len(path)-1]

	store, err := storage.OpenAt(vfs.databasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	for _, tagName := range tagNames {
		tag, err := store.Db.TagByName(tagName)
		if err != nil {
			log.Fatal(err)
		}
		if tag == nil {
			log.Fatalf("Could not retrieve tag '%v'.", tagName)
		}

		err = store.RemoveExplicitFileTag(fileId, tag.Id)
		if err != nil {
			log.Fatal(err)
		}
	}

	return fuse.OK
}

func (vfs FuseVfs) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	switch name {
	case "":
		return vfs.topDirectories()
	case "tags":
		return vfs.tagDirectories()
	}

	path := vfs.splitPath(name)
	switch path[0] {
	case "tags":
		return vfs.openTaggedEntryDir(path[1:])
	}

	return nil, fuse.ENOENT
}

func (vfs FuseVfs) Readlink(name string, context *fuse.Context) (string, fuse.Status) {
	path := vfs.splitPath(name)
	switch path[0] {
	case "tags":
		return vfs.readTaggedEntryLink(path[1:])
	}

	return "", fuse.ENOENT
}

func (vfs FuseVfs) Open(name string, flags uint32, context *fuse.Context) (fuse.File, fuse.Status) {
	data := []byte(fmt.Sprintf("db=%v\n", vfs.databasePath))
	file := fuse.NewDataFile([]byte(data))

	return file, 0
}

// implementation

func (vfs FuseVfs) splitPath(path string) []string {
	return strings.Split(path, string(filepath.Separator))
}

func (vfs FuseVfs) parseFileId(name string) (uint, error) {
	parts := strings.Split(name, ".")
	count := len(parts)

	if count == 1 {
		return 0, nil
	}

	id, err := Atoui(parts[count-2])
	if err != nil {
		id, err = Atoui(parts[count-1])
		if err != nil {
			return 0, nil
		}
	}

	return id, nil
}

func (vfs FuseVfs) topDirectories() ([]fuse.DirEntry, fuse.Status) {
	entries := make([]fuse.DirEntry, 0, 1)
	entries = append(entries, fuse.DirEntry{Name: "tags", Mode: fuse.S_IFDIR})
	return entries, fuse.OK
}

func (vfs FuseVfs) tagDirectories() ([]fuse.DirEntry, fuse.Status) {
	store, err := storage.OpenAt(vfs.databasePath)
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	defer store.Close()

	tags, err := store.Db.Tags()
	if err != nil {
		log.Fatalf("Could not retrieve tags: %v", err)
	}

	entries := make([]fuse.DirEntry, len(tags))
	for index, tag := range tags {
		entries[index] = fuse.DirEntry{Name: tag.Name, Mode: fuse.S_IFDIR}
	}

	return entries, fuse.OK
}

func (vfs FuseVfs) getTagsAttr() (*fuse.Attr, fuse.Status) {
	store, err := storage.OpenAt(vfs.databasePath)
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	defer store.Close()

	tagCount, err := store.Db.TagCount()
	if err != nil {
		log.Fatalf("Could not get tag count: %v", err)
	}

	now := time.Now()
	return &fuse.Attr{Mode: fuse.S_IFDIR | 0755, Size: uint64(tagCount), Mtime: uint64(now.Unix()), Mtimensec: uint32(now.Nanosecond())}, fuse.OK
}

func (vfs FuseVfs) getTaggedEntryAttr(path []string) (*fuse.Attr, fuse.Status) {
	pathLength := len(path)
	name := path[pathLength-1]

	store, err := storage.OpenAt(vfs.databasePath)
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	defer store.Close()

	fileId, err := vfs.parseFileId(name)
	if err != nil {
		return nil, fuse.ENOENT
	}

	if fileId == 0 {
		// tag directory

		tagIds, err := vfs.tagNamesToIds(store, path)
		if err != nil {
			log.Fatalf("Could not lookup tag IDs: %v.", err)
		}
		if tagIds == nil {
			return nil, fuse.ENOENT
		}

		fileCount, err := store.FileCountWithTags(tagIds)
		if err != nil {
			log.Fatalf("Could not retrieve count of files with tags: %v. (%v)", path, err)
		}

		now := time.Now()
		return &fuse.Attr{Mode: fuse.S_IFDIR | 0755, Size: uint64(fileCount), Mtime: uint64(now.Unix()), Mtimensec: uint32(now.Nanosecond())}, fuse.OK
	}

	file, err := store.File(fileId)
	if err != nil {
		log.Fatalf("Could not retrieve file #%v: %v", fileId, err)
	}
	if file == nil {
		return &fuse.Attr{Mode: fuse.S_IFREG}, fuse.ENOENT
	}

	fileInfo, err := os.Stat(file.Path())
	var size int64
	var modTime time.Time
	if err == nil {
		size = fileInfo.Size()
		modTime = fileInfo.ModTime()
	} else {
		size = 0
		modTime = time.Time{}
	}

	return &fuse.Attr{Mode: fuse.S_IFLNK | 0755, Size: uint64(size), Mtime: uint64(modTime.Unix()), Mtimensec: uint32(modTime.Nanosecond())}, fuse.OK
}

func (vfs FuseVfs) openTaggedEntryDir(path []string) ([]fuse.DirEntry, fuse.Status) {
	store, err := storage.OpenAt(vfs.databasePath)
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	defer store.Close()

	tagIds, err := vfs.tagNamesToIds(store, path)
	if err != nil {
		log.Fatalf("Could not lookup tag IDs: %v.", err)
	}
	if tagIds == nil {
		return nil, fuse.ENOENT
	}

	furtherTagIds, err := store.TagsForTags(tagIds)
	if err != nil {
		log.Fatalf("Could not retrieve tags for tags: %v", err)
	}

	files, err := store.FilesWithTags(tagIds, []uint{})
	if err != nil {
		log.Fatalf("Could not retrieve tagged files: %v", err)
	}

	entries := make([]fuse.DirEntry, 0, len(files)+len(furtherTagIds))
	for _, tag := range furtherTagIds {
		entries = append(entries, fuse.DirEntry{Name: tag.Name, Mode: fuse.S_IFDIR | 0755})
	}
	for _, file := range files {
		linkName := vfs.getLinkName(file)
		entries = append(entries, fuse.DirEntry{Name: linkName, Mode: fuse.S_IFLNK})
	}

	return entries, fuse.OK
}

func (vfs FuseVfs) readTaggedEntryLink(path []string) (string, fuse.Status) {
	name := path[len(path)-1]

	store, err := storage.OpenAt(vfs.databasePath)
	if err != nil {
		log.Fatalf("Could not open database: %v", err)
	}
	defer store.Close()

	fileId, err := vfs.parseFileId(name)
	if err != nil {
		log.Fatalf("Could not parse file identifier: %v", err)
	}
	if fileId == 0 {
		return "", fuse.ENOENT
	}

	file, err := store.File(fileId)
	if err != nil {
		log.Fatalf("Could not find file %v in database.", fileId)
	}

	return file.Path(), fuse.OK
}

func (vfs FuseVfs) getLinkName(file *database.File) string {
	extension := filepath.Ext(file.Path())
	fileName := filepath.Base(file.Path())
	linkName := fileName[0 : len(fileName)-len(extension)]
	suffix := "." + Uitoa(file.Id) + extension

	if len(linkName)+len(suffix) > 255 {
		linkName = linkName[0 : 255-len(suffix)]
	}

	return linkName + suffix
}

func (vfs FuseVfs) tagNamesToIds(store *storage.Storage, tagNames []string) ([]uint, error) {
	tagIds := make([]uint, len(tagNames))

	for index, tagName := range tagNames {
		tag, err := store.Db.TagByName(tagName)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve tag '%v': %v", tagName, err)
		}
		if tag == nil {
			return nil, nil
		}

		tagIds[index] = tag.Id
	}

	return tagIds, nil
}

func Uitoa(ui uint) string {
	return strconv.FormatUint(uint64(ui), 10)
}

func Atoui(str string) (uint, error) {
	ui64, err := strconv.ParseUint(str, 10, 0)
	return uint(ui64), err
}
