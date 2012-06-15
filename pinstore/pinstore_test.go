// vim:ts=4:sw=4:noexpandtab
// © 2012 Michael Stapelberg (see also: LICENSE)
//
// Testcases for the pinstore package.
package pinstore

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestPinstoreLoad(t *testing.T) {
	// Write a pin test file
	tempfile, temperr := ioutil.TempFile("/tmp/", "pinstore_test")
	if temperr != nil {
		t.Fatal("Could not create temporary file")
	}
	defer os.Remove(tempfile.Name())

	fmt.Fprintf(tempfile, `[
	{"pin":"590023", "handle":"secure"},
	{"handle":"meh", "pin": "992211"}
	]`)

	store, err := Load(tempfile.Name())
	if err != nil {
		t.Fatal("Could not create pinstore object:", err)
	}

	if val, ok := store.Pins["590023"]; !ok || val != "secure" {
		t.Error(`Pin for "secure" not found`)
	}

	http.HandleFunc("/pins", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `[{"handle":"revoked?", "pin":"1"}]`)
	})

	go http.ListenAndServe("localhost:8099", nil)

	// XXX: ugly: delay to wait until ListenAndServe actually bound the port
	time.Sleep(25 * time.Millisecond)

	store.Update("http://localhost:8099/pins")

	if val, ok := store.Pins["1"]; !ok || val != "revoked?" {
		t.Fatal("New pin not found after updating")
	}

	if _, ok := store.Pins["590023"]; ok {
		t.Fatal("Old pin still in pinstore after updating")
	}
}
