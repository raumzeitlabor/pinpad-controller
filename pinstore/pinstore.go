// vim:ts=4:sw=4
// © 2012 Michael Stapelberg (see also: LICENSE)
//
// Data structures, serialization and synchronization functions for the Pins.
package pinstore

import (
	"bytes"
	"encoding/json"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"syscall"
    "fmt"
    "time"
    "pinpad-controller/frontend"
)

var syncFailIndicatorRunning = false
var lastSyncState bool = false
var lastChecksum []byte

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

func indicateSyncFail(fe *frontend.Frontend) {
    if (syncFailIndicatorRunning == true) {
        return
    }
    syncFailIndicatorRunning = true
    for {
        if (lastSyncState == true) {
            syncFailIndicatorRunning = false
            return;
        }
        fe.LED(2, 1000)
        fe.Beep(2)
        time.Sleep(2 * time.Second)
    }
}

// Safely updates the pinstore contents with the contents from 'url'.
func (ps *Pinstore) Update(url string, fe *frontend.Frontend) (err error) {
    fmt.Printf("pinstore: trying to sync PINs\n")

	resp, err := http.Get(url)
	if err != nil {
        fmt.Printf("pinstore: could not sync PINs: %s\n", err)
        lastSyncState = false
        go indicateSyncFail(fe)
		return
	}
	defer resp.Body.Close()

	// To save some I/O (we’re on a SD card!) we calculate the hash of the
	// contents while we read them and then discard the update in case it has
	// the same content.
	checksum := crc32.NewIEEE()
	teeReader := io.TeeReader(resp.Body, checksum)

	body, err := ioutil.ReadAll(teeReader)
	if err != nil {
        fmt.Printf("pinstore: could not sync PINs: %s\n", err)
        lastSyncState = false
        go indicateSyncFail(fe)
        return
	}

	if bytes.Compare(checksum.Sum(nil), lastChecksum) == 0 {
		return
	}

	lastChecksum = checksum.Sum(nil)

	log.Printf("PINs changed, new CRC32: %x", lastChecksum)

	// Save the new pins to a new file
	file, err := ioutil.TempFile(path.Dir(ps.filename), path.Base(ps.filename)+".new")
	if err != nil {
        fmt.Printf("pinstore: could not get tmpfile: %s\n", err)
        lastSyncState = false
        go indicateSyncFail(fe)
		return
	}

	if _, err = file.Write(body); err != nil {
        fmt.Printf("pinstore: could not write PINs to tmpfile: %s\n", err)
        lastSyncState = false
        go indicateSyncFail(fe)
		return
	}

	// Try to load the new file and copy the pins over if successful
	newStore, err := Load(file.Name())
	if err != nil {
        fmt.Printf("pinstore: could not parse PINs: %s\n", err)
        lastSyncState = false
        go indicateSyncFail(fe)
		return
	}

	ps.Pins = newStore.Pins

	// Then rename the new file to the old name
	err = os.Rename(file.Name(), ps.filename)
	if err != nil {
        fmt.Printf("pinstore: could not make new PINs effective: %s\n", err)
        lastSyncState = false
        go indicateSyncFail(fe)
		return
	}

    lastSyncState = true
    fmt.Printf("pinstore: pinsync successful\n")

	return
}
