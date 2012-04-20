// vim:ts=4:sw=4:noexpandtab
// © 2012 Michael Stapelberg (see also: LICENSE)
//
// Testcases for the frontend package.
package frontend

import (
	"fmt"
	"os"
	"testing"
	"pinpad-controller/frontend"
	"time"
	"strings"
	"bytes"
)

type TestFrontend struct {
	os.File

	buffer []byte
	currentIdx int
	newBuffer chan bool
}

// Initializes a new TestFrontend instance
func NewTestFrontend() (tf *TestFrontend) {
	tf = new(TestFrontend)
	tf.newBuffer = make(chan bool)
	tf.currentIdx = 999
	return tf
}

// Puts the given string in the buffer. This buffer will be read out by Read().
func (tf *TestFrontend) FillBuffer(content string) {
	tf.buffer = []byte(content)
	tf.newBuffer <- true
}

// Returns one character from the buffer, or blocks in case the buffer is
// empty.
func (tf *TestFrontend) Read(p []byte) (n int, err error) {
	//fmt.Printf("currentIdx = %d, len = %d\n", tf.currentIdx, len(tf.buffer))
	if tf.currentIdx >= len(tf.buffer) {
		// Block until we get a new buffer
		<- tf.newBuffer
		tf.currentIdx = 0
		//fmt.Printf("after block: currentIdx = %d, len = %d\n", tf.currentIdx, len(tf.buffer))
	}
	p[0] = tf.buffer[tf.currentIdx]
	//fmt.Printf("returning byte %c\n", p[0])
	tf.currentIdx += 1
	return 1, nil
}

func (tf *TestFrontend) Write(b []byte) (n int, err error) {
	packet := string(b)
	//fmt.Printf("Write, len = %d, str = %s\n", len(b), string(b))
	if strings.HasPrefix(packet, "^PING ") {
		response := fmt.Sprintf("^PONG %c%c$", b[6], b[7])
		tf.FillBuffer(response)
	}
	return 0, nil
}

// Verify that keypresses are read properly.
func TestKeypresses(t *testing.T) {
	testfe := NewTestFrontend()
	frontend := frontend.OpenFrontendish(testfe)
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
