// vim:ts=4:sw=4:noexpandtab
package main

import (
	"fmt"
	"log"
	"flag"
	"time"
	"pinpad-controller/frontend"
	"pinpad-controller/pinpad"
	"pinpad-controller/pinstore"
	"pinpad-controller/hometec"
	"pinpad-controller/ctrlsocket"
	"pinpad-controller/tuerstatus"
	mqtt "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

var pin_url *string = flag.String(
	"pin_url",
	"http://infra.rzl/BenutzerDB/pins/haupttuer",
	"URL to load the PINs from")

var pin_path *string = flag.String(
	"pin_path",
	"/perm/pins.json",
	"Path to store the PINs permanently")

var broker = flag.String(
	"broker",
	"tcp://infra.rzl:1883",
	"The mqtt server to connect to")

var topic = flag.String(
	"topic",
	"/service/status",
	"The topic to which the door state will be published")

var lastPublishedStatus tuerstatus.Tuerstatus
var newStatus tuerstatus.Tuerstatus

// Wir haben folgende Bestandteile:
// 1) Das Frontend
// 2) Die Pin-Synchronisierung
//    Ein http-req auf die BenutzerDB.
// 3) Die Hausbus-API für SSH-öffnen
//    Separates Programm, welches dann via Hausbus den /open_door anspricht
// 4) Sensoren lesen (türstatus)
//    Hier müssen wir nur auf das RZL-Jail POSTen (vorerst).
//    Wo läuft die rrdtool-db? am besten auf dem jail zukünftig
//    Am besten stellt man den Türstatus via Hausbus-API bereit und kann das
//    dann verbasteln.
//    Kann man inotify auf /sys machen mit den GPIOs?

func updatePins(pins *pinstore.Pinstore, fe *frontend.Frontend) {
	for {
		time.Sleep(1 * time.Minute)
		pins.Update(*pin_url, fe)
	}
}

func publishMqtt() {
	opts := mqtt.NewClientOptions()
	opts.SetBroker(*broker)
	opts.SetClientId("pinpad-main")
	opts.SetCleanSession(true)
	opts.SetTraceLevel(mqtt.Off)

	opts.SetOnConnectionLost(func(client *mqtt.MqttClient, err error) {
		fmt.Printf("lost mqtt connection, trying to reconnect: %s\n", err)
		client.Start()
	})

	client := mqtt.NewClient(opts)
	_, err := client.Start()

	if err != nil {
		fmt.Printf("could not connect to mqtt broker: %s\n", err)
		return
	}

	var msg string
	if (newStatus.Open) {
		msg = "\"open\""
	} else {
		msg = "\"closed\""
	}

	mqttMsg := mqtt.NewMessage([]byte(msg))
	mqttMsg.SetQoS(mqtt.QOS_ONE)
	mqttMsg.SetRetainedFlag(true)
	r := client.PublishMessage(*topic, mqttMsg)
	<-r
	lastPublishedStatus = newStatus
	client.ForceDisconnect()
}

func main() {
	flag.Parse()

	fe, _ := frontend.OpenFrontend("/dev/ttyAMA0")
	if e := fe.Beep(frontend.BEEP_SHORT); e != nil {
		fmt.Println("cannot beep")
	}

	hometec, _ := hometec.OpenHometec()
	tuerstatusChannel := make(chan tuerstatus.Tuerstatus)
	go tuerstatus.TuerstatusPoll(tuerstatusChannel, 250 * time.Millisecond)
	go func() {
		for {
			newStatus = <-tuerstatusChannel
			if newStatus.Open {
				fe.LcdSet(" \nOpen")
			} else {
				fe.LcdSet(" \nClosed")
			}
		}
	}()

	// Ensure last door state is published. For example, if door state was
	// changed during netsplit, we need to ensure that the newest state
	// will be published as soon as network is up again.
	go func() {
		for {
			if (newStatus.Open && ! lastPublishedStatus.Open) {
				publishMqtt()
			} else if (! newStatus.Open && lastPublishedStatus.Open) {
				publishMqtt()
			}
			time.Sleep(250 * time.Millisecond)
		}
	}()

	pins, err := pinstore.Load(*pin_path)
	if err != nil {
		log.Fatalf("Could not load pins: %v", err)
	}
	if err := pins.Update(*pin_url, fe); err != nil {
		fmt.Printf("Cannot update pins: %v\n", err)
	}

	go updatePins(pins, fe)

	ctrlsocket.Listen(fe, hometec.Control)
	pinpad.ValidatePin(pins, fe, hometec.Control)
}
