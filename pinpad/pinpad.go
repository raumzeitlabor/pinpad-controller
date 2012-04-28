// vim:ts=4:sw=4:noexpandtab
package pinpad

import (
	"bytes"
	"fmt"
	"pinpad-controller/frontend"
	"pinpad-controller/pinstore"
	"regexp"
)

// Valid Pins consist of numbers only
var validPin, _ = regexp.Compile("^[0-9]+$")

// Reads keypresses from the specified frontend, verifies entered pins using
// the given pinstore and sends open/close commands to the specified hometec
// channel.
func ValidatePin(ps *pinstore.Pinstore, fe *frontend.Frontend, ht chan string) {
	var keypressBuffer bytes.Buffer
	for {
		keypress := <-fe.Keypresses
		b := []byte(keypress.Key)
		if b[0] != byte('#') {
			keypressBuffer.WriteByte(b[0])
			continue
		}

		pin := keypressBuffer.String()
		keypressBuffer.Reset()

		if pin == "666" {
			fmt.Printf("Close pin\n")
			ht <- "close"
			continue
		}

		if len(pin) != 6 || !validPin.Match([]byte(pin)) {
			fmt.Printf("Invalid PIN: %s\n", pin)
			continue
		}

		// The pin is complete, let’s validate it.
		fmt.Printf("got pin: %s\n", pin)
		if handle, ok := ps.Pins[pin]; ok {
			fmt.Printf("Successful login from %s\n", handle)
			ht <- "open"
			continue
		}

		fmt.Printf("No such PIN: %s\n", pin)
	}
}
