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

func OpenDatabase(path string) (*Database, error) {
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

// tags

func (this *Database) Tags() ([]*Tag, error) {
    sql := `SELECT id, name FROM tag`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    tags := make([]*Tag, 0, 10)
    for statement.Next() {
        var id int
        var name string
        statement.Scan(&id, &name)

        tags = append(tags, &Tag{ uint(id), name })
    }

    return tags, nil
}

func (this *Database) TagByName(name string) (*Tag, error) {
    sql := `SELECT id FROM tag WHERE name = ?`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    error = statement.Exec(name)
    if error != nil { return nil, error }
    if !statement.Next() { return nil, nil }

    var id int
    statement.Scan(&id)

    return &Tag{ uint(id), name }, nil
}

func (this *Database) TagsByFile(fileId uint) ([]*Tag, error) {
    sql := `SELECT id, name
            FROM tag
            WHERE id IN (
                          SELECT tag_id
                          FROM file_tag
                          WHERE file_id = ?
                        )`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    error = statement.Exec(int(fileId))
    if error != nil { return nil, error }

    tags := make([]*Tag, 0, 10)
    for statement.Next() {
        var tagId int
        var tagName string
        statement.Scan(&tagId, &tagName)

        tags = append(tags, &Tag{ uint(tagId), tagName })
    }

    return tags, error
}

func (this *Database) AddTag(name string) (*Tag, error) {
    sql := `INSERT INTO tag (name) VALUES (?)`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    error = statement.Exec(name)
    if error != nil { return nil, error }
    statement.Next()

    id := this.connection.LastInsertRowId()

    return &Tag{ uint(id), name }, nil
}

func (this *Database) RenameTag(tagId uint, name string) (*Tag, error) {
    sql := `UPDATE tag SET name = ? WHERE id = ?`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    error = statement.Exec(name, int(tagId))
    if error != nil { return nil, error }
    statement.Next()

    return &Tag{ tagId, name }, nil
}

func (this *Database) DeleteTag(tagId uint) (error) {
    sql := `DELETE FROM tag WHERE id = ?`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return error }
    defer statement.Finalize()

    error = statement.Exec(int(tagId))
    if error != nil { return error }
    statement.Next()

    return nil
}

// files

func (this *Database) FileByFingerprint(fingerprint string) (*File, error) {
    sql := `SELECT id FROM file WHERE fingerprint = ?`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    error = statement.Exec(fingerprint)
    if error != nil { return nil, error }
    if !statement.Next() { return nil, nil }

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

// file-paths

func (this *Database) FilePathById(filePathId uint) (*FilePath, error) {
    sql := `SELECT file_id, path
            FROM file_path
            WHERE id = ?`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    error = statement.Exec(filePathId)
    if error != nil { return nil, error }
    if !statement.Next() { return nil, nil }

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
    if !statement.Next() { return nil, nil }

    var id int
    var fileId int
    statement.Scan(&id, &fileId)

    return &FilePath{ uint(id), uint(fileId), path }, nil
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

// file-tags

func (this *Database) FileTagByFileAndTag(fileId uint, tagId uint) (*FileTag, error) {
    sql := `SELECT id FROM file_tag WHERE file_id = ? AND tag_id = ?`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    error = statement.Exec(int(fileId), int(tagId))
    if error != nil { return nil, error }
    if !statement.Next() { return nil, nil }

    var fileTagId int
    statement.Scan(&fileTagId)

    return &FileTag{ uint(fileTagId), fileId, tagId }, nil
}

func (this *Database) AddFileTag(fileId uint, tagId uint) (*FileTag, error) {
    sql := `INSERT INTO file_tag (file_id, tag_id) VALUES (?, ?)`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return nil, error }
    defer statement.Finalize()

    error = statement.Exec(int(fileId), int(tagId))
    if error != nil { return nil, error }
    if !statement.Next() { return nil, nil }

    id := this.connection.LastInsertRowId()

    return &FileTag{ uint(id), fileId, tagId }, nil
}

func (this *Database) MigrateFileTags(oldTagId uint, newTagId uint) error {
    sql := `UPDATE file_tag SET tag_id = ? WHERE tag_id = ?`

    statement, error := this.connection.Prepare(sql)
    if error != nil { return error }
    defer statement.Finalize()

    error = statement.Exec(int(newTagId), int(oldTagId))
    if error != nil { return error }
    statement.Next()

    return nil
}
