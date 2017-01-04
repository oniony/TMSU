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

package entities

type Setting struct {
	Name  string
	Value string
}

type Settings []*Setting

func (settings Settings) Len() int {
	return len(settings)
}

func (settings Settings) Less(i, j int) bool {
	return settings[i].Name < settings[j].Name
}

func (settings Settings) Swap(i, j int) {
	settings[i], settings[j] = settings[j], settings[i]
}

func (settings Settings) AutoCreateTags() bool {
	return settings.BoolValue("autoCreateTags")
}

func (settings Settings) AutoCreateValues() bool {
	return settings.BoolValue("autoCreateValues")
}

func (settings Settings) FileFingerprintAlgorithm() string {
	return settings.Value("fileFingerprintAlgorithm")
}

func (settings Settings) DirectoryFingerprintAlgorithm() string {
	return settings.Value("directoryFingerprintAlgorithm")
}

func (settings Settings) SymlinkFingerprintAlgorithm() string {
	return settings.Value("symlinkFingerprintAlgorithm")
}

func (settings Settings) ReportDuplicates() bool {
	return settings.BoolValue("reportDuplicates")
}

func (settings Settings) ContainsName(name string) bool {
	for _, setting := range settings {
		if setting.Name == name {
			return true
		}
	}

	return false
}

func (settings Settings) Value(name string) string {
	for _, setting := range settings {
		if setting.Name == name {
			return setting.Value
		}
	}

	return ""
}

func (settings Settings) BoolValue(name string) bool {
	for _, setting := range settings {
		if setting.Name == name {
			switch setting.Value {
			case "yes", "Yes", "YES", "true", "True", "TRUE":
				return true
			case "no", "No", "false", "False", "FALSE":
				return false
			default:
				panic("invalid boolean value")
			}
		}
	}

	return false
}
