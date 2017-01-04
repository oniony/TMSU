// Copyright 2011-2017 Paul Ruane.

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
	"bytes"
	"strconv"
)

type SqlBuilder struct {
	sql             bytes.Buffer
	params          []interface{}
	paramIndex      int
	needsParamComma bool
}

func NewBuilder() *SqlBuilder {
	builder := SqlBuilder{bytes.Buffer{}, make([]interface{}, 0), 1, false}
	return &builder
}

func (builder *SqlBuilder) Sql() string {
	return builder.sql.String()
}

func (builder *SqlBuilder) Params() []interface{} {
	return builder.params
}

func (builder *SqlBuilder) AppendSql(sql string) {
	if sql == "" {
		return
	}

	switch sql[0] {
	case ' ', '\n':
		// do nowt
	default:
		builder.sql.WriteRune('\n')
	}

	builder.sql.WriteString(sql)

	builder.needsParamComma = false
}

func (builder *SqlBuilder) AppendParam(value interface{}) {
	if builder.needsParamComma {
		builder.sql.WriteRune(',')
	}

	builder.sql.WriteRune('?')
	builder.sql.WriteString(strconv.Itoa(builder.paramIndex))
	builder.paramIndex++

	builder.params = append(builder.params, value)
	builder.needsParamComma = true
}
