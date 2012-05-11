// vim:ts=4:sw=4:noexpandtab
// © 2012 Michael Stapelberg (see also: LICENSE)
//
package tuerstatus

import (
	"fmt"
	"os"
	"time"
)

type Tuerstatus struct {
	// Sensor in der Tür, der zurückgibt, ob offen oder zu ist
	Open bool
}

// Runs forever and sends a byte on the given channel with the value of the
// GPIO (either '1' or '0') whenever it changes.
func gpioPoll(gpio int, output chan byte, delay time.Duration) {
	path := fmt.Sprintf("/sys/class/gpio/gpio%d/value", gpio)
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Printf("Could not open %s: %s\n", path, err)
		os.Exit(1)
	}
	value := make([]byte, 1)
	var oldValue byte = '?'
	// Repeatedly read until we have the value we were looking for
	for {
		f.Seek(0, 0)
		f.Read(value)
		if value[0] != oldValue {
			output <- value[0]
			oldValue = value[0]
		}
		time.Sleep(delay)
	}
}

// Polls the various sensors and writes an aggregated status to the channel
func TuerstatusPoll(tuerstatus chan Tuerstatus, delay time.Duration) {
	gpioValues := make(chan byte)
	go gpioPoll(25, gpioValues, delay)
	for {
		newValue := <-gpioValues
		var newStatus Tuerstatus
		newStatus.Open = (newValue == '1')
		tuerstatus <- newStatus
	}
}
