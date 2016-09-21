package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/adammck/dynamixel/iface"
	network1 "github.com/adammck/dynamixel/protocol/v1"
	network2 "github.com/adammck/dynamixel/protocol/v2"
	"github.com/adammck/dynamixel/servo"
	"github.com/adammck/dynamixel/servo/ax"
	"github.com/adammck/dynamixel/servo/xl"
	"github.com/jacobsa/go-serial/serial"
)

var (
	portName = flag.String("port", "/dev/tty.usbserial-A9ITPZVR", "the serial port path")
	servoId  = flag.Int("id", 1, "the ID of the servo to move")
	model    = flag.String("model", "ax", "the model of the servo to move")
	position = flag.Int("position", 512, "the goal position to set")
	debug    = flag.Bool("debug", false, "show serial traffic")
)

func main() {
	flag.Parse()

	options := serial.OpenOptions{
		PortName:              *portName,
		BaudRate:              1000000,
		DataBits:              8,
		StopBits:              1,
		MinimumReadSize:       0,
		InterCharacterTimeout: 100,
	}

	serial, err := serial.Open(options)
	if err != nil {
		fmt.Printf("open error: %s\n", err)
		os.Exit(1)
	}

	var nw iface.Networker
	var servo *servo.Servo

	switch *model {
	case "ax":
		nw = network1.New(serial)
		servo, err = ax.New(nw, *servoId)

	case "xl":
		nw = network2.New(serial)
		servo, err = xl.New(nw, *servoId)

	default:
		fmt.Printf("unsupported servo model: %s\n", *model)
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("servo init error: %s\n", err)
		os.Exit(1)
	}

	if *debug {
		nw.SetLogger(log.New(os.Stderr, "", log.LstdFlags))
	}

	err = servo.Ping()
	if err != nil {
		fmt.Printf("ping error: %s\n", err)
		os.Exit(1)
	}

	err = servo.SetGoalPosition(*position)
	if err != nil {
		fmt.Printf("move error: %s\n", err)
		os.Exit(1)
	}
}
