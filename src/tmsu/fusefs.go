package main

import (
    "path/filepath"
    "log"
    "os"
    "strings"
    "strconv"
    "github.com/hanwen/go-fuse/fuse"
)

type FuseVfs struct {
    fuse.DefaultFileSystem
    state *fuse.MountState
}

func MountVfs(path string) (*FuseVfs, os.Error) {
    fuseVfs := FuseVfs{}

    state, _, error := fuse.MountPathFileSystem(path, &fuseVfs, nil)
    if error != nil { return nil, error }

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

    if name == "" { return &os.FileInfo{ Mode: fuse.S_IFDIR | 0755 }, fuse.OK }

    path := strings.Split(name, string(filepath.Separator))
    log.Printf(" GetAttr(%v): path[0] = '%v'", name, path[0])

    switch (path[0]) {
        case "tags": return getTaggedEntryAttr(path[1:])
        case "untagged": return getUntaggedEntryAttr(path)
        case "view": return getViewEntryAttr(path)
    }

    log.Printf(" GetAttr(%v): unknown entry", name)

    return nil, fuse.ENOENT
}

func (this *FuseVfs) OpenDir(name string, context *fuse.Context) (chan fuse.DirEntry, fuse.Status) {
    log.Printf(">OpenDir(%v)", name)
    defer log.Printf("<OpenDir(%v)", name)

    switch name {
        case "": return topDirectories()
        case "view": return viewDirectories()
        case "tags": return tagDirectories()
        case "untagged": return untaggedFiles()
    }

    path := strings.Split(name, string(filepath.Separator))
    log.Printf(" OpenDir(%v): path[0] = '%v'", name, path[0])

    switch (path[0]) {
        case "tags": return getTaggedEntryAttr(path[1:])
        case "untagged": return getUntaggedEntryAttr(path)
        case "view": return getViewEntryAttr(path)
    }

    return nil, fuse.ENOENT
}

func (this *FuseVfs) Open(name string, flags uint32, context *fuse.Context) (fuse.File, fuse.Status) {
    log.Printf(">Open(%v)", name)
    defer log.Printf("<OpenDir(%v)", name)

    if name != "file.txt" { return nil, fuse.ENOENT }

    if flags & fuse.O_ANYWRITE != 0 { return nil, fuse.EPERM }

    return fuse.NewDataFile([]byte(name)), fuse.OK
}

// implementation

func topDirectories() (chan fuse.DirEntry, fuse.Status) {
    log.Printf(">topDirectories()")
    defer log.Printf("<topDirectories()")

    channel := make(chan fuse.DirEntry, 3)
    channel <- fuse.DirEntry{ Name: "tags", Mode: fuse.S_IFDIR }
    channel <- fuse.DirEntry{ Name: "untagged", Mode: fuse.S_IFDIR }
    channel <- fuse.DirEntry{ Name: "view", Mode: fuse.S_IFDIR }
    close(channel)

    return channel, fuse.OK
}

func viewDirectories() (chan fuse.DirEntry, fuse.Status) {
    log.Printf(">viewDirectories()")
    defer log.Printf("<viewDirectories()")

    channel := make(chan fuse.DirEntry, 0)
    //TODO dynamic view
    close(channel)

    return channel, fuse.OK
}

func tagDirectories() (chan fuse.DirEntry, fuse.Status) {
    log.Printf(">tagDirectories()")
    defer log.Printf("<tagDirectories()")

    db, error := OpenDatabase(DatabasePath())
    if error != nil { log.Fatal("Could not open database: %v", error.String()) }
    defer db.Close()

    tags, error := db.Tags()
    if error != nil { log.Fatal("Could not retrieve tags: %v", error.String()) }

    channel := make(chan fuse.DirEntry, len(tags))
    for _, tag := range tags {
        channel <- fuse.DirEntry{ Name: tag.Name, Mode: fuse.S_IFREG }
    }
    close(channel)

    return channel, fuse.OK
}

func untaggedFiles() (chan fuse.DirEntry, fuse.Status) {
    log.Printf(">untaggedFiles()")
    defer log.Printf("<untaggedFiles()")

    channel := make(chan fuse.DirEntry, 1)
    //TODO query db
    close(channel)

    return channel, fuse.OK
}

func parseFilePathId(name string) (uint, os.Error) {
    // ORIGINAL FILENAME | STORED AS             | ID
    // ------------------+-----------------------+----
    // somefile          | somefile.123          | 123
    // somefile.ext      | somefile.123.ext      | 123
    // somefile.blah     | somefile.blah.123     | 123
    // somefile.blah.ext | somefile.blah.123.ext | 123
    // somefile.456      | somefile.123.456      | 123
    // somefile.777.888  | somefile.777.123.888  | 123

    parts := strings.Split(name, ".")
    count := len(parts)

    if count == 1 { return 0, nil }

    id, error := strconv.Atoui(parts[count - 2])
    if error != nil { id, error = strconv.Atoui(parts[count - 1]) }
    if error != nil { return 0, error }

    return id, nil
}

func getTaggedEntryAttr(path []string) (*os.FileInfo, fuse.Status) {
    log.Printf(">getTaggedEntryAttr(%v)", path)
    defer log.Printf("<getTaggedEntryAttr(%v)", path)

    pathLength := len(path)
    if pathLength == 0 { return &os.FileInfo{ Mode: fuse.S_IFDIR | 0755 }, fuse.OK }

    name := path[pathLength - 1]

    log.Printf(" getTaggedEntryAttr(%v): name '%v'", path, name)

    filePathId, error := parseFilePathId(name)
    if error != nil { log.Fatalf("Could not parse file-path identifier: %v", error) }

    log.Printf(" getTaggedEntryAtry(%v): id %v", path, filePathId)

    if filePathId == 0 { return &os.FileInfo{ Mode: fuse.S_IFDIR | 0755 }, fuse.OK }

    return &os.FileInfo{ Mode: fuse.S_IFDIR | 0755 }, fuse.OK
}

func getUntaggedEntryAttr(path []string) (*os.FileInfo, fuse.Status) {
    log.Printf(">getUntaggedEntryAttr(%v)", path)
    defer log.Printf("<getUntaggedEntryAttr(%v)", path)

    if len(path) == 0 { return &os.FileInfo{ Mode: fuse.S_IFDIR | 0755 }, fuse.OK }

    return &os.FileInfo{ Mode: fuse.S_IFDIR | 0755 }, fuse.OK
}

func getViewEntryAttr(path []string) (*os.FileInfo, fuse.Status) {
    log.Printf(">getViewEntryAttr(%v)", path)
    defer log.Printf("<getViewEntryAttr(%v)", path)

    if len(path) == 0 { return &os.FileInfo{ Mode: fuse.S_IFDIR | 0755 }, fuse.OK }

    return &os.FileInfo{ Mode: fuse.S_IFDIR | 0755 }, fuse.OK
}
