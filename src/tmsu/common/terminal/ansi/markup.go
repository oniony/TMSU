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

package ansi

import (
	"strings"
)

func ParseMarkup(text string) string {
	text = strings.Replace(text, "$BOLD", string(BoldCode), -1)
	text = strings.Replace(text, "$YELLOW", string(YellowCode), -1)
	text = strings.Replace(text, "$GREEN", string(GreenCode), -1)
	text = strings.Replace(text, "$CYAN", string(CyanCode), -1)
	text = strings.Replace(text, "$WHITE", string(WhiteCode), -1)
	text = strings.Replace(text, "$RESET", string(ResetCode), -1)

	return string(text)
}
