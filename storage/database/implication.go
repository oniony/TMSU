// 2011-2015 Paul Ruane.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package database

import (
	"database/sql"
	"github.com/oniony/TMSU/entities"
)

// Retrieves the complete set of tag implications.
func Implications(tx *Tx) (entities.Implications, error) {
	sql := `
SELECT tag.id, tag.name,
       value.id, value.name,
	   implied_tag.id, implied_tag.name,
	   implied_value.id, implied_value.name
FROM implication
INNER JOIN tag tag ON implication.tag_id = tag.id
LEFT OUTER JOIN value value ON implication.value_id = value.id
INNER JOIN tag implied_tag ON implication.implied_tag_id = implied_tag.id
LEFT OUTER JOIN value implied_value ON implication.implied_value_id = implied_value.id
ORDER BY tag.name, value.name, implied_tag.name, implied_value.name`

	rows, err := tx.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	implications, err := readImplications(rows, make(entities.Implications, 0, 10))
	if err != nil {
		return nil, err
	}

	return implications, nil
}

// Retrieves the set of implications by the specified tag and value pairs.
func ImplicationsFor(tx *Tx, pairs entities.TagIdValueIdPairs) (entities.Implications, error) {
	builder := NewBuilder()

	builder.AppendSql(`
SELECT tag.id, tag.name,
       value.id, value.name,
       implied_tag.id, implied_tag.name,
       implied_value.id, implied_value.name
FROM implication
INNER JOIN tag tag ON implication.tag_id = tag.id
LEFT OUTER JOIN value value ON implication.value_id = value.id
INNER JOIN tag implied_tag ON implication.implied_tag_id = implied_tag.id
LEFT OUTER JOIN value implied_value ON implication.implied_value_id = implied_value.id
WHERE `)

	for index, pair := range pairs {
		if index > 0 {
			builder.AppendSql("   OR ")
		}

		builder.AppendSql("(implication.tag_id = ")
		builder.AppendParam(pair.TagId)
		builder.AppendSql(" AND implication.value_id IN (0, ")
		builder.AppendParam(pair.ValueId)
		builder.AppendSql("))")
	}

	builder.AppendSql(`ORDER BY tag.name, value.name, implied_tag.name, implied_value.name`)

	rows, err := tx.Query(builder.Sql(), builder.Params()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	implications, err := readImplications(rows, make(entities.Implications, 0, 10))
	if err != nil {
		return nil, err
	}

	return implications, nil
}

func ImplyingImplications(tx *Tx, pairs entities.TagIdValueIdPairs) (entities.Implications, error) {
	sql := `
SELECT tag.id, tag.name,
       value.id, value.name,
       implying_tag.id, implying_tag.name,
       implying_value.id, implying_value.name
FROM implication
INNER JOIN tag tag ON implication.tag_id = tag.id
LEFT OUTER JOIN value value ON implication.value_id = value.id
INNER JOIN tag implying_tag ON implication.implying_tag_id = implying_tag.id
LEFT OUTER JOIN value implying_value ON implication.implying_value_id = implying_value.id
WHERE `

	params := make([]interface{}, len(pairs)*2)
	for index, pair := range pairs {
		if index > 0 {
			sql += "   OR "
		}

		sql += "(implication.implied_tag_id = ? AND implication.implied_value_id = ?)"

		params[index*2] = pair.TagId
		params[index*2+1] = pair.ValueId
	}

	sql += `
ORDER BY tag.name, value.name, implying_tag.name, implying_value.name`

	rows, err := tx.Query(sql, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	implications, err := readImplications(rows, make(entities.Implications, 0, 10))
	if err != nil {
		return nil, err
	}

	return implications, nil
}

// Adds the specified implications
func AddImplication(tx *Tx, pair, impliedPair entities.TagIdValueIdPair) error {
	sql := `
INSERT OR IGNORE INTO implication (tag_id, value_id, implied_tag_id, implied_value_id)
VALUES (?1, ?2, ?3, ?4)`

	_, err := tx.Exec(sql, pair.TagId, pair.ValueId, impliedPair.TagId, impliedPair.ValueId)
	if err != nil {
		return err
	}

	return nil
}

// Deletes the specified implication
func DeleteImplication(tx *Tx, pair, impliedPair entities.TagIdValueIdPair) error {
	sql := `
DELETE FROM implication
WHERE tag_id = ?1 AND
      value_id = ?2 AND
      implied_tag_id = ?3 AND
      implied_value_id = ?4`

	result, err := tx.Exec(sql, pair.TagId, pair.ValueId, impliedPair.TagId, impliedPair.ValueId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return NoSuchImplicationError{pair, impliedPair}
	}
	if rowsAffected > 1 {
		panic("expected exactly one row to be affected")
	}

	return nil
}

// Deletes implications for the specified tag id
func DeleteImplicationsByTagId(tx *Tx, tagId entities.TagId) error {
	sql := `
DELETE FROM implication
WHERE tag_id = ?1 OR implied_tag_id = ?1`

	_, err := tx.Exec(sql, tagId)
	if err != nil {
		return err
	}

	return nil
}

// Deletes implications for the specified value id
func DeleteImplicationsByValueId(tx *Tx, valueId entities.ValueId) error {
	sql := `
DELETE FROM implication
WHERE value_id = ?1 OR implied_value_id = ?1`

	_, err := tx.Exec(sql, valueId)
	if err != nil {
		return err
	}

	return nil
}

// unexported

func readImplication(rows *sql.Rows) (*entities.Implication, error) {
	if !rows.Next() {
		return nil, nil
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var implyingTagId entities.TagId
	var implyingTagName string
	var implyingValueId *entities.ValueId
	var implyingValueName *string
	var impliedTagId entities.TagId
	var impliedTagName string
	var impliedValueId *entities.ValueId
	var impliedValueName *string
	err := rows.Scan(&implyingTagId,
		&implyingTagName,
		&implyingValueId,
		&implyingValueName,
		&impliedTagId,
		&impliedTagName,
		&impliedValueId,
		&impliedValueName)
	if err != nil {
		return nil, err
	}

	var implyingValue entities.Value
	if implyingValueId != nil {
		implyingValue = entities.Value{*implyingValueId, *implyingValueName}
	}

	var impliedValue entities.Value
	if impliedValueId != nil {
		impliedValue = entities.Value{*impliedValueId, *impliedValueName}
	}

	return &entities.Implication{entities.Tag{implyingTagId, implyingTagName},
		implyingValue,
		entities.Tag{impliedTagId, impliedTagName},
		impliedValue}, nil
}

func readImplications(rows *sql.Rows, implications entities.Implications) (entities.Implications, error) {
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
