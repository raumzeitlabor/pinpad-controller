// vim:ts=4:sw=4:noexpandtab
// © 2012 Michael Stapelberg (see also: LICENSE)
//
// This package implements the protocol to speak with the hometec and provides
// high-level methods.
package hometec

import (
	"fmt"
	"os"
	"time"
)

type Hometec struct {
	Control chan string
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
		fmt.Printf("Could not open /sys/class/gpio/export: %s\n", err)
		os.Exit(1)
	}
	f.Write([]byte(fmt.Sprintf("%s\n", direction)))
	f.Close()
}

func gpioSet(gpio int, value int) {
	path := fmt.Sprintf("/sys/class/gpio/gpio%d/value", gpio)
	f, err := os.OpenFile(path, os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Could not open /sys/class/gpio/export: %s\n", err)
		os.Exit(1)
	}
	f.Write([]byte(fmt.Sprintf("%d\n", value)))
	f.Close()
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
	// Den Dreh-Motor starten, dann 50ms warten, damit er auch läuft.
	aufdrehenStarten()
	time.Sleep(50 * time.Millisecond)

	// Für 100ms einkoppeln.
	einkoppelnStarten()
	time.Sleep(100 * time.Millisecond)
	einkoppelnStoppen()

	// Nun dreht der Motor den Schlüssel.
	// TODO: solange drehen, bis offen ist, nicht immer 2 sekunden
	time.Sleep(2 * time.Second)
	aufdrehenStoppen()
	time.Sleep(50 * time.Millisecond)

	// Jetzt für 100ms auskoppeln.
	auskoppelnStarten()
	time.Sleep(100 * time.Millisecond)
	auskoppelnStoppen()
}

func (hometec *Hometec) Close() {
	// Den Dreh-Motor starten, dann 50ms warten, damit er auch läuft.
	zudrehenStarten()
	time.Sleep(50 * time.Millisecond)

	// Für 100ms einkoppeln.
	einkoppelnStarten()
	time.Sleep(100 * time.Millisecond)
	einkoppelnStoppen()

	// Nun dreht der Motor den Schlüssel.
	// TODO: solange drehen, bis offen ist, nicht immer 2 sekunden
	time.Sleep(2 * time.Second)
	zudrehenStoppen()
	time.Sleep(50 * time.Millisecond)

	// Jetzt für 100ms auskoppeln.
	auskoppelnStarten()
	time.Sleep(100 * time.Millisecond)
	auskoppelnStoppen()
}
