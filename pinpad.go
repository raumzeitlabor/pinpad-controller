// vim:ts=4:sw=4:noexpandtab
package main

import (
	"fmt"
	"pinpad-controller/frontend"
)

func main() {
	fe, _ := frontend.OpenFrontend("/dev/ttyUSB0")
	if e := fe.Beep(frontend.BEEP_SHORT); e != nil {
		fmt.Println("cannot beep")
	}

	for {
		select {
		case keypress := <- fe.Keypresses:
			fmt.Printf("Got keypress: %s\n", keypress.Key)
		}
	}
}
