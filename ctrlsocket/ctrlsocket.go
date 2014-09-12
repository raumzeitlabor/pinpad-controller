// vim:ts=4:sw=4:noexpandtab
package ctrlsocket

import (
	"os"
    "net"
	"fmt"
	"pinpad-controller/frontend"
)

func Listen(fe *frontend.Frontend, ht chan string) {
    _ = os.Remove("/tmp/pinpad-ctrl.sock")
    l, err := net.Listen("unix", "/tmp/pinpad-ctrl.sock")
    if err != nil {
        fmt.Printf("pinpad-ctrl: listen error: %s\n", err)
        return
    }

    go func() {
        for {
            fd, err := l.Accept()
            if err != nil {
                fmt.Printf("pinpad-ctrl: accept error: %s\n", err)
                return
            }
            go cmdHandler(fd, fe, ht)
        }
    }()
}

func cmdHandler(c net.Conn, fe *frontend.Frontend, ht chan string) {
    for {
        buf := make([]byte, 32)
        nr, err := c.Read(buf)
        if err != nil {
            fmt.Printf("pinpad-ctrl: could not read from sock: %s\n", err);
            return
        }

        if (buf[nr-1] == '\n') {
            nr--
        }

        var resp []byte
        data := buf[0:nr]
        fmt.Printf("pinpad-ctrl: read: %s\n", string(data))
        switch string(data) {
            case "open":
                ht <- "open"
                resp = []byte("ok\n")
            case "close":
                ht <- "close"
                resp = []byte("ok\n")
            default:
                resp = []byte("error: unknown cmd\n")
        }

        // confirm cmd
        _, err = c.Write(resp)
        if err != nil {
            fmt.Printf("pinpad-ctrl: could not write to sock: %s\n", err)
        }
    }
}
