// vim:ts=4:sw=4:noexpandtab
package pinpad

import (
	"bytes"
	"fmt"
	"pinpad-controller/frontend"
	"pinpad-controller/pinstore"
	"pinpad-controller/tuerstatus"
	"regexp"
	"time"
)

// Valid Pins consist of numbers only
var validPin, _ = regexp.Compile("^[0-9]+$")

func invalidPin(pin string, fe *frontend.Frontend) {
	fmt.Printf("Invalid PIN: %s\n", pin)
	fe.LcdSet("Invalid PIN!")
	fe.LED(2, 3000)
	go func() {
		time.Sleep(2 * time.Second)
		if tuerstatus.CurrentStatus().Open {
			fe.LcdSet(" \nOpen")
		} else {
			fe.LcdSet(" \nClosed")
		}
	}()
}

// Reads keypresses from the specified frontend, verifies entered pins using
// the given pinstore and sends open/close commands to the specified hometec
// channel.
func ValidatePin(ps *pinstore.Pinstore, fe *frontend.Frontend, ht chan string) {
	var keypressBuffer bytes.Buffer
	for {
		keypress := <-fe.Keypresses
		b := []byte(keypress.Key)
		fe.LED(1, 50)
		fe.Beep(2)
		if b[0] != byte('#') {
			if keypressBuffer.Len() == 0 {
				fe.LcdSet("PIN: *")
			} else {
				fe.LcdPut("*")
			}
			keypressBuffer.WriteByte(b[0])
			continue
		}

		pin := keypressBuffer.String()
		keypressBuffer.Reset()

		if pin == "666" || pin == "*" {
			fmt.Printf("Got close pin, locking door\n")
			fe.LcdSet("Locking door...")
			fe.LED(3, 3000)
			fe.LED(2, 1)
			ht <- "close"
			continue
		}

		if len(pin) != 6 || !validPin.Match([]byte(pin)) {
			invalidPin(pin, fe)
			continue
		}

		// The pin is complete, let’s validate it.
		if handle, ok := ps.Pins[pin]; ok {
			fmt.Printf("%s unlocked the door\n", handle)
			fe.LcdSet(fmt.Sprintf("Welcome, %s!\nUnlocking door...", handle))
			fe.LED(3, 3000)
			fe.LED(2, 1)
			ht <- "open"
			continue
		}

		fmt.Printf("No such PIN: %s\n", pin)
		invalidPin(pin, fe)
	}
}
