// vim:ts=4:sw=4
// © 2012 Michael Stapelberg (see also: LICENSE)
//
// Data structures, serialization and synchronization functions for the Pins.
package pinstore

import (
	"log"
	"encoding/gob"
	"os"
	"syscall"
)

type Pinstore struct {
	filename string
	Pins map[string]string
}

func Load(filename string) (result *Pinstore) {
	result = new(Pinstore)
	result.filename = filename
	result.Pins = make(map[string]string, 0)

	// Check if the file exists and load it, if so. Otherwise, create a new file.
	if _, err := os.Stat(filename); err != nil {
		if e, ok := err.(*os.PathError); ok && e.Err != syscall.ENOENT {
			log.Fatalf(`Error loading histogram data from "%s": %s`, filename, err)
		} 

		// Err == os.ENOENT, this is ok.
	} else {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatalf(`Error reading histogram data from "%s": %s`, filename, err)
		}
		defer file.Close()

		decoder := gob.NewDecoder(file)
		if err := decoder.Decode(&result); err != nil {
			log.Fatalf(`Could not load histogram from "%s": %s`, filename, err)
		}
	}

	return result
}
