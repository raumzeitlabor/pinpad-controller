// vim:ts=4:sw=4:noexpandtab
// © 2012 Michael Stapelberg (see also: LICENSE)
//
// Unfortunately, Go has no package to use a tty interface, which we use to
// communicate with the Pinpad frontend via RS232. Therefore, we need to use
// the low-level syscall wrappers here.
package uart

import (
	"os"
	"syscall"
	"unsafe"
	"errors"
	"io"
)

type TTY struct {
	os.File
}

type TTYish interface {
	io.Reader
	io.Writer
}

// termios types
type cc_t byte
type speed_t uint
type tcflag_t uint

// speed constants
const (
	B0 = iota
	B50
	B75
	B110
	B134
	B150
	B200
	B300
	B600
	B1200
	B1800
	B2400
	B4800
	B9600
	B19200
	B38400
	B57600 = 0010001
	B115200
)

// termios constants
const (
    BRKINT = tcflag_t (0000002);
    ICRNL = tcflag_t (0000400);
    INPCK = tcflag_t (0000020);
    ISTRIP = tcflag_t (0000040);
    IXON = tcflag_t (0002000);
    OPOST = tcflag_t (0000001);
    CS8 = tcflag_t (0000060);
    ECHO = tcflag_t (0000010);
    ICANON = tcflag_t (0000002);
    IEXTEN = tcflag_t (0100000);
    ISIG = tcflag_t (0000001);
    CBAUD = tcflag_t (0010017);
    CIBAUD = tcflag_t (002003600000);
    VTIME = tcflag_t (5);
    VMIN = tcflag_t (6);
)

const (
	TIOCM_RTS = 0x004
)

const NCCS = 32
type termios struct {
    c_iflag, c_oflag, c_cflag, c_lflag tcflag_t;
    c_line cc_t;
    c_cc [NCCS]cc_t;
    c_ispeed, c_ospeed speed_t
}

// ioctl constants
const (
	TCGETS = 0x5401
	TCSETS = 0x5402
	TCSETSW = 0x5403
	TCSETSF = 0x5404
	TIOCMGET = 0x5415
	TIOCMSET = 0x5418
)

var fd uintptr

func (tty *TTY) setTermios(src *termios) error {
	fd := tty.Fd()
	r1, _, _ := syscall.RawSyscall(syscall.SYS_IOCTL,
                                     uintptr(fd), uintptr(TCSETSF),
                                     uintptr(unsafe.Pointer(src)));

    if r1 != 0 {
        return errors.New("Error in TCSETS")
    }

    return nil
}

// We don’t need this right now using the usb2serial. Maybe on the Raspberry Pi
// itself? We’ll see.
//// Enables/Disables RTS (Ready To Send)
//func (tty *TTY) SetRTS(rts uint) error {
//	var mcs uint
//	fd := tty.Fd()
//	r1, _, _ := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(TIOCMGET), uintptr(unsafe.Pointer(&mcs)))
//	if r1 != 0 {
//		return errors.New("Error in TIOCMGET")
//	}
//	mcs = mcs | TIOCM_RTS
//	r1, _, _ = syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(TIOCMSET), uintptr(unsafe.Pointer(&mcs)))
//	if r1 != 0 {
//		return errors.New("Error in TIOCMSET")
//	}
//
//	return nil
//}

func OpenTTY(path string, speed speed_t) (tty *TTY, err error) {
	uartFile, e := os.OpenFile(path, os.O_RDWR | syscall.O_NOCTTY, 0)
	if e != nil {
		return nil, e
	}

	uartTTY := &TTY{*uartFile}

	var newState termios
	newState.c_iflag = 0
	newState.c_oflag = 0
	newState.c_lflag = 0
	newState.c_cflag = syscall.CLOCAL | syscall.CREAD | syscall.CS8 | tcflag_t(speed)
	// VMIN = 1 means that at least one character needs to be in the input
	// buffer before read() returns. This makes us able to use blocking reads.
	newState.c_cc[VMIN] = 1

	if e := uartTTY.setTermios(&newState); e != nil {
		return nil, e
	}

	return uartTTY, nil
}
