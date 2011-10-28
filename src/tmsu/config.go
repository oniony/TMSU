package main

import (
    "os"
    "path/filepath"
)

const defaultDatabaseName = ".tmsu/db"

func DatabasePath() string {
    path, error := os.Getenverror("TMSU_DB")
    if (error == nil) { return path }

    homePath, error := os.Getenverror("HOME")
    if (error != nil) { panic("No home directory.") }

    return filepath.Join(homePath, defaultDatabaseName)
}
