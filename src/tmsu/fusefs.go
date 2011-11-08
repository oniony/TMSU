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

    path string
    state *fuse.MountState
}

func MountVfs(path string) (*FuseVfs, error) {
    fuseVfs := FuseVfs{}

    state, _, error := fuse.MountPathFileSystem(path, &fuseVfs, nil)
    if error != nil { return nil, error }

    fuseVfs.path = path
    fuseVfs.state = state

    return &fuseVfs, nil
}

func (this *FuseVfs) Unmount() {
    this.state.Unmount()
}

func (this *FuseVfs) Loop() {
    this.state.Loop()
}

func (this *FuseVfs) GetAttr(name string, context *fuse.Context) (*os.FileInfo, fuse.Status) {
    log.Printf(">GetAttr(%v)", name)
    defer log.Printf("<GetAttr(%v)", name)

    switch name {
        case "": fallthrough
        case "tags": return &os.FileInfo{ Mode: fuse.S_IFDIR | 0755 }, fuse.OK
    }

    path := splitPath(name)
    switch (path[0]) {
        case "tags": return getTaggedEntryAttr(path[1:])
    }

    return nil, fuse.ENOENT
}

func (this *FuseVfs) OpenDir(name string, context *fuse.Context) (chan fuse.DirEntry, fuse.Status) {
    log.Printf(">OpenDir(%v)", name)
    defer log.Printf("<OpenDir(%v)", name)

    switch name {
        case "": return topDirectories()
        case "tags": return tagDirectories()
    }

    path := splitPath(name)
    switch (path[0]) {
        case "tags": return openTaggedEntryDir(path[1:])
    }

    return nil, fuse.ENOENT
}

func (this *FuseVfs) Open(name string, flags uint32, context *fuse.Context) (fuse.File, fuse.Status) {
    log.Printf(">Open(%v)", name)
    defer log.Printf("<Open(%v)", name)

    //if flags & fuse.O_ANYWRITE != 0 { return nil, fuse.EPERM }

    return fuse.NewDataFile([]byte("tmsu (c) 2011 Paul Ruane\n")), fuse.OK
}

func (this *FuseVfs) Readlink(name string, context *fuse.Context) (string, fuse.Status) {
    log.Printf(">Readlink(%v)", name)
    defer log.Printf("<Readlink(%v)", name)

    path := splitPath(name)
    switch (path[0]) {
        case "tags": return readTaggedEntryLink(path[1:])
    }

    return "", fuse.ENOENT
}

// implementation

func splitPath(path string) []string {
    return strings.Split(path, string(filepath.Separator))
}

func parseFilePathId(name string) (uint, error) {
    log.Printf(">parseFilePathId(%v)", name)
    defer log.Printf("<parseFilePathId(%v)", name)

    parts := strings.Split(name, ".")
    count := len(parts)

    if count == 1 { return 0, nil }

    id, error := strconv.Atoui(parts[count - 2])
    if error != nil { id, error = strconv.Atoui(parts[count - 1]) }
    if error != nil { return 0, error }

    log.Printf(" parseFilePathId(%v): %v", name, id)

    return id, nil
}

func topDirectories() (chan fuse.DirEntry, fuse.Status) {
    log.Printf(">topDirectories()")
    defer log.Printf("<topDirectories()")

    channel := make(chan fuse.DirEntry, 2)
    channel <- fuse.DirEntry{ Name: "tags", Mode: fuse.S_IFDIR }
    close(channel)

    return channel, fuse.OK
}

func tagDirectories() (chan fuse.DirEntry, fuse.Status) {
    log.Printf(">tagDirectories()")
    defer log.Printf("<tagDirectories()")

    db, error := OpenDatabase(databasePath())
    if error != nil { log.Fatal("Could not open database: %v", error) }
    defer db.Close()

    tags, error := db.Tags()
    if error != nil { log.Fatal("Could not retrieve tags: %v", error) }

    channel := make(chan fuse.DirEntry, len(tags))
    for _, tag := range tags {
        channel <- fuse.DirEntry{ Name: tag.Name, Mode: fuse.S_IFREG }
    }
    close(channel)

    return channel, fuse.OK
}

func getTaggedEntryAttr(path []string) (*os.FileInfo, fuse.Status) {
    log.Printf(">getTaggedEntryAttr(%v)", path)
    defer log.Printf("<getTaggedEntryAttr(%v)", path)

    pathLength := len(path)
    name := path[pathLength - 1]

    log.Printf(" getTaggedEntryAttr(%v): name '%v'", path, name)

    filePathId, error := parseFilePathId(name)
    if error != nil { log.Fatalf("Could not parse file-path identifier: %v", error) }

    log.Printf(" getTaggedEntryAttr(%v): id %v", path, filePathId)

    if filePathId == 0 { return &os.FileInfo{ Mode: fuse.S_IFDIR | 0755 }, fuse.OK }

    return &os.FileInfo{ Mode: fuse.S_IFLNK | 0755, Size: int64(10) }, fuse.OK
}

func openTaggedEntryDir(path []string) (chan fuse.DirEntry, fuse.Status) {
    log.Printf(">openTaggedEntryDir(%v)", path)
    defer log.Printf("<openTaggedEntryDir(%v)", path)

    db, error := OpenDatabase(databasePath())
    if error != nil { log.Fatalf("Could not open database: %v", error) }
    defer db.Close()

    //TODO assumption that all path dirs are tags

    filePaths, error := db.FilePathsByTag(path)
    if error != nil { log.Fatalf("Could not retrieve tagged files: %v", error) }

    channel := make(chan fuse.DirEntry, len(filePaths))
    for _, filePath := range filePaths {
        channel <- fuse.DirEntry { Name: "file." + strconv.Uitoa(filePath.Id), Mode: fuse.S_IFLNK }
    }
    close(channel)

    return channel, fuse.OK
}

func readTaggedEntryLink(path []string) (string, fuse.Status) {
    log.Printf(">readTaggedEntryLink(%v)", path)
    defer log.Printf("<readTaggedEntryLink(%v)", path)

    name := path[len(path) - 1]

    log.Printf(" readTaggedEntryLink(%v): name '%v'", path, name)

    filePathId, error := parseFilePathId(name)
    if error != nil { log.Fatalf("Could not parse file-path identifier: %v", error) }

    log.Printf(" readTaggedEntryLink(%v): id %v", path, filePathId)

    if filePathId == 0 { return "", fuse.ENOENT }

    db, error := OpenDatabase(databasePath())
    if error != nil { log.Fatalf("Could not open database: %v", error) }
    defer db.Close()

    filePath, error := db.FilePathById(filePathId)
    if error != nil { log.Fatalf("Could not find file-path %v in database.", filePathId) }

    return filePath.Path, fuse.OK
}
