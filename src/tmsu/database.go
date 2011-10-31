package main

import (
    "fmt"
    "os"
    "gosqlite.googlecode.com/hg/sqlite"
)

type Database struct {
    connection  *sqlite.Conn
}

func OpenDatabase(path string) (*Database, os.Error) {
    connection, error := sqlite.Open(path)

    if (error != nil) {
        fmt.Fprintf(os.Stderr, "Could not open database: %v.", error)
        return nil, error
    }

    database := Database{connection}

    return &database, nil
}

func (this *Database) Close() {
    this.connection.Close()
}

func (this *Database) Tags() ([]Tag, os.Error) {
    statement, error := this.connection.Prepare("SELECT * FROM tag")

    if (error != nil) { return nil, error }

    tags := make([]Tag, 0, 10)

    for statement.Next() {
        var rowId int
        var tag string
        statement.Scan(&rowId, &tag)

        tags = append(tags, Tag{rowId, tag})
    }

    return tags, nil
}

func (this *Database) Tagged(tag string) ([]string, os.Error) {
    return nil, nil
}
