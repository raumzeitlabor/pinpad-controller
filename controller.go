// vim:ts=4:sw=4:noexpandtab
package main

import (
	"fmt"
	"time"
	"pinpad-controller/frontend"
	"pinpad-controller/pinpad"
	"pinpad-controller/pinstore"
	"pinpad-controller/hometec"
	"pinpad-controller/tuerstatus"
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
	pins.Pins["112233"] = "secure"

	fe, _ := frontend.OpenFrontend("/dev/ttyAMA0")
	if e := fe.Beep(frontend.BEEP_SHORT); e != nil {
		fmt.Println("cannot beep")
	}

	hometec, _ := hometec.OpenHometec()
	tuerstatusChannel := make(chan tuerstatus.Tuerstatus)
	go tuerstatus.TuerstatusPoll(tuerstatusChannel, 250 * time.Millisecond)
	go func() {
		for {
			newStatus := <-tuerstatusChannel
			if newStatus.Open {
				fe.LcdSet("Open")
			} else {
				fe.LcdSet("Closed")
			}
		}
	}()
	pinpad.ValidatePin(pins, fe, hometec.Control)
}
