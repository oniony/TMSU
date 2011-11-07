package main

import (
    "fmt"
    "os"
    "strconv"
    "strings"
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
    sql := "SELECT id, name FROM tag"

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }

    tags := make([]Tag, 0, 10)
    for statement.Next() {
        var id int
        var tag string
        statement.Scan(&id, &tag)

        tags = append(tags, Tag{ id, tag })
    }

    return tags, nil
}

func (this *Database) Tagged(tagNames []string) ([]FilePath, os.Error) {
    sql := `SELECT id, file_id, path
            FROM file_path
            WHERE id IN (
                            SELECT file_path_id
                            FROM file_path_tag
                            WHERE tag_id IN (
                                                SELECT id
                                                FROM tag
                                                WHERE name IN (` + strings.Repeat("?,", len(tagNames) - 1) + `?)
                                            )
                            GROUP BY file_path_id
                            HAVING count(*) = ` + strconv.Itoa(len(tagNames)) + `
                        )`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }

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

func (this *Database) File(fingerprint string) (*File, os.Error) {
    sql := "SELECT id FROM file WHERE fingerprint = ?"

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }

    error = statement.Exec(fingerprint)
    if error != nil { return nil, error }
    if !statement.Next() { return nil, nil }

    var id int
    statement.Scan(&id)

    return &File{id, fingerprint}, nil
}

func (this *Database) FilePath(filePathId uint) (*FilePath, os.Error) {
    sql := `SELECT file_id, path
            FROM file_path
            WHERE id = ?`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }

    error = statement.Exec(filePathId)
    if error != nil { return nil, error }
    if !statement.Next() { return nil, nil }

    var fileId int
    var path string
    statement.Scan(&fileId, &path)

    return &FilePath{ filePathId, uint(fileId), path }, nil
}
