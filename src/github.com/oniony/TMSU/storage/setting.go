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

package storage

import (
	"github.com/oniony/TMSU/entities"
	"github.com/oniony/TMSU/storage/database"
	"sort"
)

var defaultSettings = entities.Settings{
	&entities.Setting{"autoCreateTags", "yes"},
	&entities.Setting{"autoCreateValues", "yes"},
	&entities.Setting{"directoryFingerprintAlgorithm", "none"},
	&entities.Setting{"fileFingerprintAlgorithm", "dynamic:SHA256"},
	&entities.Setting{"reportDuplicates", "yes"},
	&entities.Setting{"symlinkFingerprintAlgorithm", "follow"}}

// The complete set of settings.
func (storage *Storage) Settings(tx *Tx) (entities.Settings, error) {
	settings, err := database.Settings(tx.tx)
	if err != nil {
		return nil, err
	}

	// enrich with defaults
	for _, defaultSetting := range defaultSettings {
		if !settings.ContainsName(defaultSetting.Name) {
			settings = append(settings, defaultSetting)
		}
	}

	sort.Sort(settings)

	return settings, nil
}

func (storage *Storage) Setting(tx *Tx, name string) (*entities.Setting, error) {
	setting, err := database.Setting(tx.tx, name)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		value := defaultSettings.Value(name)
		setting = &entities.Setting{name, value}
	}

	return setting, nil
}

func (storage *Storage) UpdateSetting(tx *Tx, name, value string) (*entities.Setting, error) {
	return database.UpdateSetting(tx.tx, name, value)
}
