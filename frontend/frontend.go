// vim:ts=4:sw=4:noexpandtab
// © 2012 Michael Stapelberg (see also: LICENSE)
//
// This package implements the protocol to speak with the frontend and provides
// high-level methods.
package frontend

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"pinpad-controller/uart"
	"strings"
	"time"
)

type beepkind uint

const (
	BEEP_LONG  = beepkind(0)
	BEEP_SHORT = beepkind(1)
)

type Frontend struct {
	tty        uart.TTYish
	Keypresses chan KeyPressEvent
}

type KeyPressEvent struct {
	// Which key was pressed? Can be '0' to '9', '*' or '#'.
	Key string
}

func OpenFrontendish(ttyish uart.TTYish) *Frontend {
	fe := new(Frontend)
	fe.Keypresses = make(chan KeyPressEvent)
	fe.tty = ttyish
	go fe.readAndPing()
	return fe
}

func OpenFrontend(path string) (frontend *Frontend, err error) {
	ttyish, e := uart.OpenTTY(path, uart.B9600)
	if e != nil {
		return nil, e
	}
	doorStateWatcher()
	return OpenFrontendish(ttyish), nil
}

func doorStateWatcher() {
// Check door state and provide debug info
	go func() {
		for {
			time.Sleep(1 * time.Second)
			//fmt.Printf("doorStateWatcher watches the door's state\n")
		}
	}()
}

// readAndPing takes care of reading bytes, filling a buffer and then sending
// the message on the communication channel. Also, it triggers a PING request
// every second.
//
// readAndPing is called as a Go function in OpenFrontend.
func (fe *Frontend) readAndPing() {
	// We have one Go function running in the background and triggering a
	// message every second. Upon receiving the message, readAndPing() will
	// send a PING request to the frontend.
	secondPassed := make(chan bool, 1)
	go func() {
		for {
			time.Sleep(1 * time.Second)
			secondPassed <- true
		}
	}()

	// We have another Go function which will (blockingly) read from the TTY.
	// Whenever there is a new byte received (which is not the 0 byte, our
	// protocol is ASCII only), it will be sent on the byteChannel.
	byteChannel := make(chan byte)
	go func() {
		reader := bufio.NewReader(fe.tty)
		for {
			nextByte, err := reader.ReadByte()
			if err != nil {
				fmt.Printf("Error reading from the serial interface: %s\n", err)
				fmt.Printf("Do you have console=ttyS0 in your kernel cmdline maybe?\n")
				os.Exit(1)
			}

			// Sometimes, we get zero bytes on the UART (for example while the
			// frontend is initializing), so we filter these.
			if nextByte == 0 {
				continue
			}

			byteChannel <- nextByte
		}
	}()

	var receiveBuffer bytes.Buffer
	// Stores the PING value we sent to the frontend. Will be checked before
	// sending the next value so that we can detect packet loss. If this is the
	// empty string, the frontend PONGed, otherwise it contains the value we
	// sent but did not get acknowledged.
	previousPing := ""
	for {
		select {
		case nextByte := <-byteChannel:
			// If we are at the beginning of a buffer, we only accept the
			// leading "^" byte.
			if receiveBuffer.Len() == 0 && nextByte != '^' {
				continue
			}
			receiveBuffer.WriteByte(nextByte)
			// Check whether the buffer is complete yet. Every buffer has
			// exactly 9 bytes.
			if receiveBuffer.Len() != 9 {
				continue
			}
			packet := receiveBuffer.String()
			if strings.HasPrefix(packet, "^PONG ") {
				pong := packet[len("^PONG ") : len("^PONG ")+2]
				if pong != previousPing {
					fmt.Printf("pinpad-frontend sent %s, but we expected %s\n", pong, previousPing)
				}
				// Clear previous ping, that means it was acknowledged
				previousPing = ""
			} else if strings.HasPrefix(packet, "^PAD ") {
				var event KeyPressEvent
				event.Key = string(receiveBuffer.Bytes()[5])
				fe.Keypresses <- event
			}
			receiveBuffer.Reset()

		case <-secondPassed:
			if previousPing != "" {
				fmt.Printf("pinpad-frontend did not PONG %s\n", previousPing)
			}
			// 58 is the amount of (printable) characters from '@' to 'z'.
			previousPing = fmt.Sprintf("%c%c", rand.Int31n(58)+'@', rand.Int31n(58)+'@')
			fe.Ping(previousPing)
		}
	}
}

func (fe *Frontend) Ping(rnd string) error {
	_, err := fe.tty.Write([]byte(fmt.Sprintf("^PING %s                             $", rnd)))
	if err != nil {
		return err
	}

	return nil
}

func (fe *Frontend) Beep(kind beepkind) error {
	command := fmt.Sprintf("^BEEP %d                              $", kind)
	_, err := fe.tty.Write([]byte(command))
	if err != nil {
		return err
	}

	return nil
}

func (fe *Frontend) LcdSet(text string) error {
	maxlength := len("^LCD $") + 32
	command := fmt.Sprintf("^LCD %s", text)
	for len(command) < maxlength {
		command = fmt.Sprintf("%s ", command)
	}
	command = fmt.Sprintf("%s$", command)
	_, err := fe.tty.Write([]byte(command))
	if err != nil {
		return err
	}

	return nil
}

func (fe *Frontend) LcdPut(char string) error {
	command := fmt.Sprintf("^LCH %s                               $", char)
	_, err := fe.tty.Write([]byte(command))
	if err != nil {
		return err
	}

	return nil
}

func (fe *Frontend) LED(idx int, duration int) error {
	maxlength := len("^LED $") + 32
	command := fmt.Sprintf("^LED %d %d", idx, duration)
	for len(command) < maxlength {
		command = fmt.Sprintf("%s ", command)
	}
	command = fmt.Sprintf("%s$", command)
	_, err := fe.tty.Write([]byte(command))
	if err != nil {
		return err
	}

	return nil
}
