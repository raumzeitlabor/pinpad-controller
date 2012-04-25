// vim:ts=4:sw=4:noexpandtab
package main

import (
	"fmt"
	"pinpad-controller/frontend"
	"pinpad-controller/pinpad"
	"pinpad-controller/pinstore"
)

// Wir haben folgende Bestandteile:
// 1) Das Frontend
// 2) Die Pin-Synchronisierung
//    Ein http-req auf die BenutzerDB.
// 3) Die Hausbus-API für SSH-öffnen
//    Separates Programm, welches dann via Hausbus den /open_door anspricht
// 4) Sensoren lesen (türstatus)
//    Hier müssen wir nur auf das RZL-Jail POSTen (vorerst).
//    Wo läuft die rrdtool-db? am besten auf dem jail zukünftig
//    Am besten stellt man den Türstatus via Hausbus-API bereit und kann das
//    dann verbasteln.
//    Kann man inotify auf /sys machen mit den GPIOs?

func main() {
	pins := pinstore.Load("/tmp/pins")
	pins.Pins["1234"] = "secure"

	fe, _ := frontend.OpenFrontend("/dev/ttyUSB0")
	if e := fe.Beep(frontend.BEEP_SHORT); e != nil {
		fmt.Println("cannot beep")
	}

	hometecControl := make(chan string)
	go func() {
		for {
			command := <-hometecControl
			switch command {
			case "open":
				fmt.Printf("should tell the hometec to open the door\n")
			}
		}
	}()
	pinpad.ValidatePin(pins, fe, hometecControl)
}
