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
	"strconv"
)

type SqlBuilder struct {
	Sql    string
	Params []interface{}

	paramIndex int
	needsComma bool
}

func NewBuilder() *SqlBuilder {
	return &SqlBuilder{"", make([]interface{}, 0), 1, false}
}

func (builder *SqlBuilder) AppendSql(sql string) {
	builder.Sql += " " + sql

	builder.needsComma = false
}

func (builder *SqlBuilder) AppendParam(value interface{}) {
	if builder.needsComma {
		builder.Sql += ","
	}

	builder.Sql += "?" + strconv.Itoa(builder.paramIndex)
	builder.paramIndex++

	builder.Params = append(builder.Params, value)

	builder.needsComma = true
}
