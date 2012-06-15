// vim:ts=4:sw=4
// © 2012 Michael Stapelberg (see also: LICENSE)
//
// Data structures, serialization and synchronization functions for the Pins.
package pinstore

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"syscall"
)

// This type is used temporarily when decoding the JSON only.
type pin struct {
	Handle string
	Pin    string
}

type Pinstore struct {
	filename string
	Pins     map[string]string
}

func Load(filename string) (*Pinstore, error) {
	result := new(Pinstore)
	result.filename = filename
	result.Pins = make(map[string]string, 0)

	// Check if the file exists and load it, if so. Otherwise, create a new file.
	if _, err := os.Stat(filename); err != nil {
		// ENOENT is okay (we create a new file), but anything else is not.
		if e, ok := err.(*os.PathError); ok && e.Err != syscall.ENOENT {
			return nil, err
		}
	} else {
		// Read the whole file into pinContents
		pinContents, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		// The file is JSON encoded, so unmarshal it into the pins array first…
		var pins []pin
		if err := json.Unmarshal(pinContents, &pins); err != nil {
			return nil, err
		}

		// …then fill the Pins map for convenience
		for _, pin := range pins {
			result.Pins[pin.Pin] = pin.Handle
		}
	}

	return result, nil
}

// Safely updates the pinstore contents with the contents from 'url'.
func (ps *Pinstore) Update(url string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// Save the new pins to a new file
	file, err := ioutil.TempFile(path.Dir(ps.filename), path.Base(ps.filename)+".new")
	if err != nil {
		return
	}
	if _, err = file.Write(body); err != nil {
		return
	}

	// Try to load the new file and copy the pins over if successful
	newStore, err := Load(file.Name())
	if err != nil {
		return
	}

	ps.Pins = newStore.Pins

	// Then rename the new file to the old name
	err = os.Rename(file.Name(), ps.filename)

	return
}
