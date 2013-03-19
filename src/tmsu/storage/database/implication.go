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

package database

import (
	"database/sql"
	"strings"
)

type Implication struct {
	ImplyingTag Tag
	ImpliedTag  Tag
}

type Implications []*Implication

// Retrieves the complete set of tag implications.
func (db *Database) Implications() (Implications, error) {
	sql := `SELECT t1.id, t1.name, t2.id, t2.name
            FROM implication, tag t1, tag t2
            WHERE implication.tag_id = t1.id
            AND implication.implied_tag_id = t2.id
            ORDER BY t1.name, t2.name`

	result, err := db.connection.Query(sql)
	if err != nil {
		return nil, err
	}

	implications, err := readImplications(result, make(Implications, 0, 10))
	if err != nil {
		return nil, err
	}

	return implications, nil
}

// Retrieves the set of tags implied by the specified tags.
func (db *Database) ImplicationsForTags(tags Tags) (Implications, error) {
	sql := `SELECT t1.id, t1.name, t2.id, t2.name
            FROM implication, tag t1, tag t2
            WHERE implication.tag_id IN (?`
	sql += strings.Repeat(",?", len(tags)-1)
	sql += `)
	        AND implication.tag_id = t1.id
	        AND implication.implied_tag_id = t2.id`

	params := make([]interface{}, len(tags))
	for index, tag := range tags {
		params[index] = tag.Id
	}

	result, err := db.connection.Query(sql, params...)
	if err != nil {
		return nil, err
	}

	implications, err := readImplications(result, make(Implications, 0, 10))
	if err != nil {
		return nil, err
	}

	return implications, nil
}

// Updates implications featuring the specified tag.
func (db Database) UpdateImplicationsForTagId(implyingTagId, impliedTagId uint) error {
	// prevent a tag implying itself

	sql := `DELETE from implication
            WHERE (tag_id = ?1 AND implied_tag_id = ?2)
            OR (tag_id = ?2 AND implied_tag_id = ?1)`

	_, err := db.connection.Exec(sql, implyingTagId, impliedTagId)
	if err != nil {
		return err
	}

	sql = `UPDATE implication
           SET tag_id = ?2
           WHERE tag_id = ?1`

	_, err = db.connection.Exec(sql, implyingTagId, impliedTagId)
	if err != nil {
		return err
	}

	sql = `UPDATE implication
           SET implied_tag_id = ?2
           WHERE implied_tag_id = ?1`

	_, err = db.connection.Exec(sql, implyingTagId, impliedTagId)
	if err != nil {
		return err
	}

	return nil
}

// Adds the specified implications
func (db Database) AddImplication(tagId, impliedTagId uint) error {
	sql := `INSERT OR IGNORE INTO implication (tag_id, implied_tag_id)
	        VALUES (?1, ?2)`

	_, err := db.connection.Exec(sql, tagId, impliedTagId)
	if err != nil {
		return err
	}

	return nil
}

// Deletes the specified implications
func (db Database) DeleteImplication(tagId, impliedTagId uint) error {
	sql := `DELETE FROM implication
            WHERE tag_id = ?1 AND implied_tag_id = ?2`

	_, err := db.connection.Exec(sql, tagId, impliedTagId)
	if err != nil {
		return err
	}

	return nil
}

// Deletes implications featuring the specified tag.
func (db Database) DeleteImplicationsForTagId(tagId uint) error {
	sql := `DELETE FROM implication
            WHERE tag_id = ?1 OR implied_tag_id = ?1`

	_, err := db.connection.Exec(sql, tagId)
	if err != nil {
		return err
	}

	return nil
}

// unexported

func readImplication(rows *sql.Rows) (*Implication, error) {
	if !rows.Next() {
		return nil, nil
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var implyingTagId uint
	var implyingTagName string
	var impliedTagId uint
	var impliedTagName string
	err := rows.Scan(&implyingTagId, &implyingTagName, &impliedTagId, &impliedTagName)
	if err != nil {
		return nil, err
	}

	return &Implication{Tag{implyingTagId, implyingTagName}, Tag{impliedTagId, impliedTagName}}, nil
}

func readImplications(rows *sql.Rows, implications Implications) (Implications, error) {
	for {
		implication, err := readImplication(rows)
		if err != nil {
			return nil, err
		}
		if implication == nil {
			break
		}

		implications = append(implications, implication)
	}

	return implications, nil
}
