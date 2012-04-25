// vim:ts=4:sw=4:noexpandtab
// © 2012 Michael Stapelberg (see also: LICENSE)
//
// Testcases for the pinpad package.
package pinpad

import (
	"bytes"
	"fmt"
	"pinpad-controller/frontend"
	"pinpad-controller/pinstore"
	"pinpad-controller/testfrontend"
	"testing"
	"time"
)

func resultWithBuffer(testfe *testfrontend.TestFrontend, buffer string, hometec chan string) (string, bool) {
	// Trigger a timeout after 0.25s
	timeout := make(chan bool)
	go func() {
		time.Sleep(250 * time.Millisecond)
		timeout <- true
	}()

	testfe.FillBuffer(buffer)

	select {
	case str := <-hometec:
		return str, true
	case <-timeout:
		return "", false
	}

	return "", false
}

// Constructs a buffer for the TestFrontend which contains keypresses that form
// the given pin.
func constructPinBuffer(pin string) string {
	var buffer bytes.Buffer
	for _, v := range pin {
		buffer.WriteString(fmt.Sprintf("^PAD %c  $", v))
	}
	buffer.WriteString("^PAD #  $")
	return buffer.String()
}

func TestPinValidation(t *testing.T) {
	testfe := testfrontend.NewTestFrontend()
	frontend := frontend.OpenFrontendish(testfe)

	pins := pinstore.Load("/tmp/testcase")
	pins.Pins["123456"] = "secure"

	hometec := make(chan string)
	go ValidatePin(pins, frontend, hometec)

	// TODO: gofmt
	invalidPin := constructPinBuffer("1234")
	validPin := constructPinBuffer("123456")

	if _, ok := resultWithBuffer(testfe, invalidPin, hometec); ok {
		t.Error("Hometec got an instruction for an invalid pin")
	}

	if _, ok := resultWithBuffer(testfe, validPin, hometec); ok {
		t.Error("Hometec got an instruction for an invalid pin")
	}
}
