software für den pinpad controller v2 (Raspberry Pi, programmiert in Go)

### Building

    apt-get install golang golang-doc
    export GOPATH=…
    GOARCH=arm go build

### Installation on Raspberry Pi

    scp systemd/* raspberry:/etc/systemd/system/
    scp pinpad-controller tuersshd /usr/local/bin/
    ssh raspberry
    # adduser tuersshd
    # systemctl enable pinpad-tuersshd.service
    # systemctl enable pinpad-fixperms.service
    # systemctl enable pinpad-controller.service
