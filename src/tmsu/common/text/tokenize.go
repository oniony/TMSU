/*
Copyright 2011-2014 Paul Ruane.

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

package text

func Tokenize(text string) []string {
    tokens := make([]string, 0, 10)
    token := make([]rune, 0, 100)
    var quote rune = 0
    escape := false

    for _, char := range text {
        switch {
        case escape:
            token = append(token, char)
            escape = false
        case char == '\\':
            escape = true
        case quote != 0:
            if char == quote {
                tokens = append(tokens, string(token))
                token = make([]rune, 0, 100)
                quote = 0
            } else {
                token = append(token, char)
            }
        case char == '"', char == '\'':
            quote = char
        case char == ' ', char == '\t':
            if len(token) > 0 {
                tokens = append(tokens, string(token))
                token = make([]rune, 0, 100)
            }
        default:
            token = append(token, char)
        }
    }

    if len(token) > 0 {
        tokens = append(tokens, string(token))
    }

    return tokens
}
