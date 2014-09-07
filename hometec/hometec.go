// vim:ts=4:sw=4:noexpandtab
// © 2012 Michael Stapelberg (see also: LICENSE)
// © 2014 Simon Elsbrock (see also: LICENSE)
//
// This package implements the protocol to speak with the hometec and provides
// high-level methods.
//
// gpio7: im türramén, 1 == abgeschlossen
// gpio8: in der tür, 1 == offen, 0 == 2x zu
// gpio24: in der tür, 1 == zu, 0 == offen
//
// haupttür hat anders als lagertür nur eine schließumdrehung.
package hometec

import (
	"fmt"
	"os"
	"time"
)

type Hometec struct {
	Control chan string
}

func gpioWaitFor(gpio int, wantedValue byte) bool {
	path := fmt.Sprintf("/sys/class/gpio/gpio%d/value", gpio)
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Printf("Could not open %s: %s\n", path, err)
		os.Exit(1)
	}
	value := make([]byte, 1)
	// Repeatedly read until we have the value we were looking for
	for {
		f.Seek(0, 0)
		f.Read(value)
		if value[0] == wantedValue {
			return true
		}
		fmt.Printf("pin %d state: %s\n", gpio, value)
		time.Sleep(250)
	}
	f.Close()
	return false
}

func gpioWaitForWithTimeout(gpio int, wantedValue byte, timeout time.Duration) bool {
	closed := make(chan bool)
	go func() {
		closed <- gpioWaitFor(gpio, wantedValue)
	}()
	go func() {
		time.Sleep(timeout)
		closed <- false
	}()
	return <-closed
}

func gpioExport(gpio int) {
	f, err := os.OpenFile("/sys/class/gpio/export", os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Could not open /sys/class/gpio/export: %s\n", err)
		os.Exit(1)
	}
	f.Write([]byte(fmt.Sprintf("%d\n", gpio)))
	f.Close()
}

func gpioSetDirection(gpio int, direction string) {
	path := fmt.Sprintf("/sys/class/gpio/gpio%d/direction", gpio)
	f, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Could not open %s: %s\n", path, err)
		os.Exit(1)
	}
	f.Write([]byte(fmt.Sprintf("%s\n", direction)))
	f.Close()
}

func gpioSet(gpio int, value int) {
	path := fmt.Sprintf("/sys/class/gpio/gpio%d/value", gpio)
	f, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Could not open %s: %s\n", path, err)
		os.Exit(1)
	}
	f.Write([]byte(fmt.Sprintf("%d\n", value)))
	f.Close()
}

func gpioGet(gpio int) []byte {
	path := fmt.Sprintf("/sys/class/gpio/gpio%d/value", gpio)
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Printf("Could not open %s: %s\n", path, err)
		os.Exit(1)
	}
	value := make([]byte, 1)
	f.Seek(0, 0)
	f.Read(value)
	f.Close()
	return value
}

func OpenHometec() (hometec *Hometec, err error) {
	hometec = new(Hometec)
	hometec.Control = make(chan string)

	// Initialize the GPIOs
	outputGPIOs := []int { 11, 9, 10, 22, 21, 17, 4, 1 };
	inputGPIOs := []int { 18, 23, 24, 8, 7, 25 };
	allGPIOs := make([]int, len(outputGPIOs) + len(inputGPIOs))
	copy(allGPIOs, outputGPIOs)
	copy(allGPIOs[len(outputGPIOs):], inputGPIOs)

	// Export all the GPIOs
	for _, gpio := range(allGPIOs) {
		gpioExport(gpio)
	}

	// Configure outputs and set them to high. The hometec uses inverted logic.
	for _, gpio := range(outputGPIOs) {
		gpioSetDirection(gpio, "out")
		gpioSet(gpio, 1)
	}

	// Configure inputs
	for _, gpio := range(inputGPIOs) {
		gpioSetDirection(gpio, "in")
	}

	// Now read the control channel and react to any commands.
	go hometec.readControlChannel()

	return hometec, nil
}

func (hometec *Hometec) readControlChannel() {
	for {
		command := <-hometec.Control
		fmt.Printf("read command: %s\n", command)
		if command == "open" {
			hometec.Open()
		}
		if command == "close" {
			hometec.Close()
		}
	}
}

// Stilbruch: Deutsche Funktionsnamen. Ist aber sinnig, damit man versteht, was
// hier geschieht.

// Startet den Einkopplungs-Motor. Dieser Motor sorgt dafür, dass der
// Haupt-motor, der dann tatsächlich den Schlüssel dreht, eingekoppelt wird.
// Wenn er nicht eingekoppelt ist, kann man von Hand am Rad drehen, also die
// Tür mit einem Schlüssel ganz normal aufschließen.
func einkoppelnStarten() {
	gpioSet(11, 0)
	gpioSet(9, 0)
	gpioSet(10, 0)
}

// Stoppt den Einkopplungs-Motor.
func einkoppelnStoppen() {
	gpioSet(11, 1)
	gpioSet(9, 1)
	gpioSet(10, 1)
}

// Motor zum Öffnen drehen
func aufdrehenStarten() {
	gpioSet(1, 0)
	gpioSet(17, 0)
	gpioSet(4, 0)
}

func aufdrehenStoppen() {
	gpioSet(1, 1)
	gpioSet(17, 1)
	gpioSet(4, 1)
}

// Motor zum Schließen drehen
func zudrehenStarten() {
	gpioSet(21, 0)
	gpioSet(17, 0)
	gpioSet(4, 0)
}

func zudrehenStoppen() {
	gpioSet(21, 1)
	gpioSet(17, 1)
	gpioSet(4, 1)
}

func auskoppelnStarten() {
	gpioSet(22, 0)
	gpioSet(9, 0)
	gpioSet(10, 0)
}

func auskoppelnStoppen() {
	gpioSet(22, 1)
	gpioSet(9, 1)
	gpioSet(10, 1)
}

func (hometec *Hometec) Open() {
	// Gar nicht erst aufschließen, wenn schon offen
	if gpioGet(8)[0] == '0' {
		return
	}

	// Den Dreh-Motor starten, dann 50ms warten, damit er auch läuft.
	aufdrehenStarten()
	time.Sleep(50 * time.Millisecond)

	// Für 100ms einkoppeln.
	einkoppelnStarten()
	time.Sleep(100 * time.Millisecond)
	einkoppelnStoppen()

	// Nun dreht der Motor den Schlüssel.
	gpioWaitForWithTimeout(8, '0', 4 * time.Second)
	// Noch eine halbe Sekunde mehr drehen, damit auch wirklich offen ist
	time.Sleep(500 * time.Millisecond)
	aufdrehenStoppen()
	time.Sleep(50 * time.Millisecond)

	// Jetzt für 100ms auskoppeln.
	auskoppelnStarten()
	time.Sleep(100 * time.Millisecond)
	auskoppelnStoppen()
}

func (hometec *Hometec) Close() {
	// Gar nicht erst zuschließen, wenn schon zu
	if gpioGet(24)[0] == '0' {
		return
	}

	// Den Dreh-Motor starten, dann 50ms warten, damit er auch läuft.
	zudrehenStarten()
	time.Sleep(50 * time.Millisecond)

	// Für 100ms einkoppeln.
	einkoppelnStarten()
	time.Sleep(100 * time.Millisecond)
	einkoppelnStoppen()

	// Nun dreht der Motor den Schlüssel.
	gpioWaitForWithTimeout(24, '0', 2 * time.Second)
	zudrehenStoppen()
	time.Sleep(50 * time.Millisecond)

	// Jetzt für 100ms auskoppeln.
	auskoppelnStarten()
	time.Sleep(100 * time.Millisecond)
	auskoppelnStoppen()
}
