package db

import (
    "fmt"
    "os"
    "gosqlite.googlecode.com/hg/sqlite"
)

func Open(path string) *Database {
    conn, err := sqlite.Open(path)

    if (err != nil) {
        fmt.Fprintf(os.Stderr, "Could not open database: %v.", err)
        return nil
    }

    database := Database{conn}

    return &database
}

type Database struct {
    connection  *sqlite.Conn
}

func (this *Database) Close() {
    this.Close()
}

// tag

// file

// file-tag

// file-path
