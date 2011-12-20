package main

import (
    "path/filepath"
    "os"
)

func isRegular(fileInfo os.FileInfo) bool {
    return fileInfo.Mode() & os.ModeType == 0
}

func directoryEntries(path string) ([]string, error) {
    file, error := os.Open(path)
    if error != nil { return nil, error }
    defer file.Close()

    entryNames, error := file.Readdirnames(0)
    if error != nil { return nil, error }

    entryPaths := make([]string, len(entryNames))
    for index, entryName := range entryNames {
        entryPaths[index] = filepath.Join(path, entryName)
    }

    return entryPaths, nil
}
