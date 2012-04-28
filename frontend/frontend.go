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
	ttyish, e := uart.OpenTTY(path, uart.B38400)
	if e != nil {
		return nil, e
	}
	return OpenFrontendish(ttyish), nil
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
				// TODO: proper error handling
				fmt.Printf("err: %s\n", err)
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
	// TODO: keep track of pings so that we can detect packet loss
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
				fmt.Println("got a PONG reply :)")
			} else if strings.HasPrefix(packet, "^PAD ") {
				var event KeyPressEvent
				event.Key = string(receiveBuffer.Bytes()[5])
				fe.Keypresses <- event
			}
			receiveBuffer.Reset()

		case <-secondPassed:
			fmt.Printf("a second passed, sending ping\n")
			// TODO: Generate random value
			fe.Ping("aa")
		}
	}
}

func (fe *Frontend) Ping(rnd string) error {
	_, err := fe.tty.Write([]byte("^PING cc                             $"))
	if err != nil {
		return err
	}

	return nil
}

func (fe *Frontend) Beep(kind beepkind) error {
	_, err := fe.tty.Write([]byte("^BEEP 1                              $"))
	if err != nil {
		return err
	}

	return nil
}

// TODO: LED-Steuerung
// TODO: LCD-Steuerung
