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

package ansi

var ResetCode string = esc + "0m"

var BoldCode string = esc + "1m"
var ItalicCode string = esc + "3m"
var UnderlineCode string = esc + "4m"
var BlinkCode string = esc + "5m"
var InvertCode string = esc + "7m"

var BlackCode string = esc + "30m"
var RedCode string = esc + "31m"
var GreenCode string = esc + "32m"
var YellowCode string = esc + "33m"
var BlueCode string = esc + "34m"
var MagentaCode string = esc + "35m"
var CyanCode string = esc + "36m"
var WhiteCode string = esc + "37m"

var DarkGreyCode string = BoldCode + BlackCode

var CodeByName = map[string]string{
	"reset":    ResetCode,
	"bold":     BoldCode,
	"italic":   ItalicCode,
	"blink":    BlinkCode,
	"invert":   InvertCode,
	"black":    BlackCode,
	"red":      RedCode,
	"green":    GreenCode,
	"yellow":   YellowCode,
	"blue":     BlueCode,
	"magenta":  MagentaCode,
	"cyan":     CyanCode,
	"white":    WhiteCode,
	"darkgrey": DarkGreyCode,
}

// unexported

var esc = "\x1b["
