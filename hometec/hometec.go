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

func (hometec *Hometec) Open() {
	gpioSet(22, 0)
	gpioSet(9, 0)
	gpioSet(10, 0)
	time.Sleep(3 * time.Second)
	gpioSet(22, 1)
	gpioSet(9, 1)
	gpioSet(10, 1)
}

func (hometec *Hometec) Close() {
}
