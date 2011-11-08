package main

import (
    "errors"
    "fmt"
    "os"
    "strconv"
    "strings"
    "gosqlite.googlecode.com/hg/sqlite"
)

var NOT_FOUND error = errors.New("No such item.")

type Database struct {
    connection  *sqlite.Conn
}

func OpenDatabase(path string) (*Database, error) {
    fmt.Println("Database", path)

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

func (this *Database) Tags() ([]Tag, error) {
    sql := `SELECT id, name FROM tag`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    tags := make([]Tag, 0, 10)
    for statement.Next() {
        var id int
        var tag string
        statement.Scan(&id, &tag)

        tags = append(tags, Tag{ id, tag })
    }

    return tags, nil
}

func (this *Database) FilePathsByTag(tagNames []string) ([]FilePath, error) {
    sql := `SELECT id, file_id, path
            FROM file_path
            WHERE file_id IN (
                SELECT file_id
                FROM file_tag
                WHERE tag_id IN (
                    SELECT id
                    FROM tag
                    WHERE name IN (` + strings.Repeat("?,", len(tagNames) - 1) + `?)
                )
                GROUP BY file_id
                HAVING count(*) = ` + strconv.Itoa(len(tagNames)) + `
            )`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    // convert string array to empty-interface array
    castTagNames := make([]interface{}, len(tagNames))
    for index, tagName := range tagNames { castTagNames[index] = tagName }

    error = statement.Exec(castTagNames...)
    if error != nil { return nil, error }

    filePaths := make([]FilePath, 0, 10)
    for statement.Next() {
        var filePathId int
        var fileId int
        var path string
        statement.Scan(&filePathId, &fileId, &path)

        filePaths = append(filePaths, FilePath{ uint(filePathId), uint(fileId), path })
    }

    return filePaths, nil
}

func (this *Database) FileByFingerprint(fingerprint string) (*File, error) {
    sql := `SELECT id FROM file WHERE fingerprint = ?`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    error = statement.Exec(fingerprint)
    if error != nil { return nil, error }
    if !statement.Next() { return nil, NOT_FOUND }

    var id int
    statement.Scan(&id)

    return &File{uint(id), fingerprint}, nil
}

func (this *Database) AddFile(fingerprint string) (*File, error) {
    sql := `INSERT INTO file (fingerprint) VALUES (?)`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    error = statement.Exec(fingerprint)
    if error != nil { return nil, error }
    statement.Next()

    id := this.connection.LastInsertRowId()

    return &File{uint(id), fingerprint}, nil
}

func (this *Database) FilePathById(filePathId uint) (*FilePath, error) {
    sql := `SELECT file_id, path
            FROM file_path
            WHERE id = ?`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    error = statement.Exec(filePathId)
    if error != nil { return nil, error }
    if !statement.Next() { return nil, NOT_FOUND }

    var fileId int
    var path string
    statement.Scan(&fileId, &path)

    return &FilePath{ filePathId, uint(fileId), path }, nil
}

func (this *Database) FilePathByPath(path string) (*FilePath, error) {
    sql := `SELECT id, file_id
            FROM file_path
            WHERE path = ?`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    error = statement.Exec(path)
    if error != nil { return nil, error }
    if !statement.Next() { return nil, NOT_FOUND }

    var id int
    var fileId int
    statement.Scan(&id, &fileId)

    return &FilePath{ uint(id), uint(fileId), path }, nil
}

func (this *Database) AddFilePath(fileId uint, path string) (*FilePath, error) {
    sql := `INSERT INTO file_path (file_id, path) VALUES (?, ?)`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    error = statement.Exec(int(fileId), path)
    if error != nil { return nil, error }
    statement.Next()

    id := this.connection.LastInsertRowId()

    return &FilePath{ uint(id), fileId, path }, nil
}

