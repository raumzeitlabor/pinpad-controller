// vim:ts=4:sw=4:noexpandtab
package main

import (
	"fmt"
	"log"
	"flag"
	"time"
	"pinpad-controller/frontend"
	"pinpad-controller/pinpad"
	"pinpad-controller/pinstore"
	"pinpad-controller/hometec"
	"pinpad-controller/tuerstatus"
)

var pin_url *string = flag.String(
	"pin_url",
	"https://blackbox.raumzeitlabor.de/BenutzerDB/pins/getraenkelager",
	"URL to load the PINs from")

var pin_path *string = flag.String(
	"pin_path",
	"/perm/pins.json",
	"Path to store the PINs permanently")

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

func updatePins(pins *pinstore.Pinstore) {
	for {
		time.Sleep(1 * time.Minute)
		pins.Update(*pin_url)
	}
}

func main() {
	flag.Parse()

	pins, err := pinstore.Load(*pin_path)
	if err != nil {
		log.Fatalf("Could not load pins: %v", err)
	}
	if err := pins.Update(*pin_url); err != nil {
		fmt.Printf("Cannot update pins: %v\n", err)
	}

	go updatePins(pins)

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
				fe.LcdSet(" \nOpen")
			} else {
				fe.LcdSet(" \nClosed")
			}
		}
	}()
	pinpad.ValidatePin(pins, fe, hometec.Control)
}
