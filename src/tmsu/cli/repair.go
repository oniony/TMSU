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

package cli

import (
	"fmt"
	"os"
	"tmsu/common/fingerprint"
	"tmsu/common/log"
	"tmsu/entities"
	"tmsu/storage"
)

var RepairCommand = Command{
	Name:     "repair",
	Aliases:  []string{"fix"},
	Synopsis: "Repair the database",
	Description: `tmsu [OPTION]... repair [PATH]...

Fixes broken paths and stale fingerprints in the database caused by file
modifications and moves.

                                          Reported Repaired
    Modified files                          yes      yes
    Moved files                             yes      yes
    Missing files                           yes      no
    Unmodified files                        no       

Modified files are identified by a change to the file's modification time or
file size. These files are repaired by updating the modification time, size and
fingerprint in the database.

Moved files will only be repaired if a file with the same fingerprint can be
found under PATHs. Files that have been both moved and modified will not be
identified.

Missing files are reported but are not, by default, removed from the database
as this would destroy the tagging information associated with them. To remove
these files use the --remove option.

Examples:

    $ tmsu repair
    $ tmsu repair /some/path /new/path
    $ tmsu repair --remove`,
	Options: Options{{"--pretend", "-P", "do not make any changes", false, ""},
		{"--remove", "-R", "remove missing files from the database", false, ""},
		{"--unmodified", "-u", "recalculate fingerprints for unmodified files", false, ""}},
	Exec: repairExec,
}

func repairExec(options Options, args []string) error {
	pretend := options.HasOption("--pretend")
	removeMissing := options.HasOption("--remove")
	recalcUnmodified := options.HasOption("--unmodified")
	searchPaths := args

	store, err := storage.Open()
	if err != nil {
		return fmt.Errorf("could not open storage: %v", err)
	}
	defer store.Close()

	if err := store.Begin(); err != nil {
		return fmt.Errorf("could not begin transaction: %v", err)
	}
	defer store.Commit()

	fingerprintAlgorithm, err := store.SettingAsString("fingerprintAlgorithm")
	if err != nil {
		return err
	}

	log.Infof(2, "retrieving all files from the database.")

	//TODO limit to path if specified

	dbFiles, err := store.Files()
	if err != nil {
		return fmt.Errorf("could not retrieve paths from storage: %v", err)
	}

	unmodfied, modified, missing := determineStatuses(dbFiles)

	if recalcUnmodified {
		if err = repairUnmodified(store, unmodfied, pretend, fingerprintAlgorithm); err != nil {
			return err
		}
	}

	if err = repairModified(store, modified, pretend, fingerprintAlgorithm); err != nil {
		return err
	}

	if err = repairMoved(store, missing, searchPaths, pretend, fingerprintAlgorithm); err != nil {
		return err
	}

	if err = repairMissing(store, missing, pretend, removeMissing); err != nil {
		return err
	}

	//TODO cleanup: any files that have no tags: remove
	//TODO cleanup: any tags that do not correspond to a file: remove

	return nil
}

// unexported

func determineStatuses(dbFiles entities.Files) (unmodified, modified, missing entities.Files) {
	log.Infof(2, "determining file statuses")

	unmodified = make(entities.Files, 0, 10)
	modified = make(entities.Files, 0, 10)
	missing = make(entities.Files, 0, 10)

	for _, dbFile := range dbFiles {
		stat, err := os.Stat(dbFile.Path())
		if err != nil {
			switch {
			case os.IsPermission(err):
				log.Warnf("%v: permission denied", dbFile.Path())
				continue
			case os.IsNotExist(err):
				missing = append(missing, dbFile.Path())
				continue
			}
		}

		if dbFile.ModTime == stat.ModTime().UTC() && dbFile.Size == stat.Size() {
			unmodified = append(unmodified, dbFile)
		} else {
			modified = append(modified, dbFile)
		}
	}

	return
}

func repairUnmodified(store *storage.Storage, unmodified entities.Files, pretend bool, fingerprintAlgorithm string) error {
	log.Infof(2, "recalculating fingerprints for unmodified files")

	for _, dbFile := range unmodified {
		stat, err := os.Stat(dbFile.Path())
		if err != nil {
			return err
		}

		fingerprint, err := fingerprint.Create(dbFile.Path(), fingerprintAlgorithm)
		if err != nil {
			log.Warnf("%v: could not create fingerprint: %v", dbFile.Path(), err)
			continue
		}

		if !pretend {
			_, err := store.UpdateFile(dbFile.Id, dbFile.Path(), fingerprint, stat.ModTime(), stat.Size(), stat.IsDir())
			if err != nil {
				return fmt.Errorf("%v: could not update file in database: %v", dbFile.Path(), err)
			}

			log.Infof(1, "%v: unmodified [FIXED]", dbFile.Path())
		} else {
			log.Infof(1, "%v: unmodified", dbFile.Path())
		}
	}

	return nil
}

func repairModified(store *storage.Storage, modified entities.Files, pretend bool, fingerprintAlgorithm string) error {
	log.Infof(2, "repairing modified files")

	for _, dbFile := range modified {
		stat, err := os.Stat(dbFile.Path())
		if err != nil {
			return err
		}

		fingerprint, err := fingerprint.Create(dbFile.Path(), fingerprintAlgorithm)
		if err != nil {
			log.Warnf("%v: could not create fingerprint: %v", dbFile.Path(), err)
			continue
		}

		if !pretend {
			_, err := store.UpdateFile(dbFile.Id, dbFile.Path(), fingerprint, stat.ModTime(), stat.Size(), stat.IsDir())
			if err != nil {
				return fmt.Errorf("%v: could not update file in database: %v", dbFile.Path(), err)
			}

			log.Infof(1, "%v: modified [FIXED]", dbFile.Path())
		} else {
			log.Infof(1, "%v: modified", dbFile.Path())
		}
	}

	return nil
}

func repairMoved(store *storage.Storage, missing entities.Files, searchPaths []string, pretend bool, fingerprintAlgorithm string) error {
	log.Infof(2, "repairing moved files")

	if len(missing) == 0 || len(searchPaths) == 0 {
		// don't bother enumerating filesystem if nothing to do
		return nil
	}

	pathsBySize := make(map[uint][]string, 10)

	//TODO enumerate search paths and build map of size -> list of paths

	//TODO for each missing
	//TODO find set with same size
	//TODO confirm match with fingerprint comparison
	//TODO nil entry in missing

	/*
			moved := make([]string, 0, 10)
			for path, dbFile := range missing {
				log.Infof(2, "%v: searching for new location", path)

				for candidatePath, stat := range untagged {
					if stat.Size() == dbFile.Size {
						fingerprint, err := fingerprint.Create(candidatePath, fingerprintAlgorithm)
						if err != nil {
							return fmt.Errorf("%v: could not create fingerprint: %v", path, err)
						}

						if fingerprint == dbFile.Fingerprint {
							moved = append(moved, path)

							if !pretend {
		                        log.Infof(1, "%v: moved to %v [FIXED]", path, candidatePath)

								_, err := store.UpdateFile(dbFile.Id, candidatePath, dbFile.Fingerprint, stat.ModTime(), dbFile.Size, dbFile.IsDir)
								if err != nil {
									return fmt.Errorf("%v: could not update file in database: %v", path, err)
								}
							} else {
		                        log.Infof(1, "%v: moved to %v", path, candidatePath)
		                    }

							delete(untagged, candidatePath)

							break
						}
					}
				}
			}

			for _, path := range moved {
				delete(missing, path)
			}
	*/

	return nil
}

func repairMissing(store *storage.Storage, missing entities.Files, pretend, force bool) error {
	for _, dbFile := range missing {
		if force && !pretend {
			if err := store.DeleteFileTagsByFileId(dbFile.Id); err != nil {
				return fmt.Errorf("%v: could not delete file-tags: %v", dbFile.Path(), err)
			}

			log.Infof(1, "%v: missing [FIXED]", dbFile.Path())
		} else {
			log.Infof(1, "%v: missing", dbFile.Path())
		}
	}

	return nil
}
