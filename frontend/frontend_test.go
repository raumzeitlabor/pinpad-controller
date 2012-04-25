// vim:ts=4:sw=4:noexpandtab
// © 2012 Michael Stapelberg (see also: LICENSE)
//
// Testcases for the frontend package.
package frontend

import (
	"bytes"
	"pinpad-controller/testfrontend"
	"testing"
	"time"
)

// Verify that keypresses are read properly.
func TestKeypresses(t *testing.T) {
	testfe := testfrontend.NewTestFrontend()
	frontend := OpenFrontendish(testfe)
	testfe.FillBuffer("^PAD 2  $^PAD 3  $^PAD 4  $^PAD 5  $^PAD #  $")

	// Trigger a timeout after 0.5s
	timeout := make(chan bool)
	go func() {
		time.Sleep(500 * time.Millisecond)
		timeout <- true
	}()

	// Now verify that we receive the key presses we put in the buffer earlier.
	var keypressBuffer bytes.Buffer

	for {
		select {
		case <-timeout:
			t.Error("Did not receive the expected keypresses within 0.5s\n")
			return
		case keypress := <-frontend.Keypresses:
			b := []byte(keypress.Key)
			keypressBuffer.WriteByte(b[0])
			if keypressBuffer.String() == "2345#" {
				return
			}
		}
	}
}
