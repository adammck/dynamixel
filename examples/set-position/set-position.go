package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/adammck/dynamixel/iface"
	"github.com/adammck/dynamixel/network"
	p1 "github.com/adammck/dynamixel/protocol/v1"
	p2 "github.com/adammck/dynamixel/protocol/v2"
	"github.com/adammck/dynamixel/servo"
	"github.com/adammck/dynamixel/servo/ax"
	"github.com/adammck/dynamixel/servo/xl"
	"github.com/jacobsa/go-serial/serial"
)

var (
	portName = flag.String("port", "/dev/tty.usbserial-A9ITPZVR", "the serial port path")
	servoID  = flag.Int("id", 1, "the ID of the servo to move")
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

	network := network.New(serial)
	if *debug {
		network.SetLogger(log.New(os.Stderr, "", log.LstdFlags))
	}

	network.Flush()

	var proto iface.Protocol
	var servo *servo.Servo

	switch *model {
	case "ax":
		proto = p1.New(network)
		servo, err = ax.New(proto, *servoID)

	case "xl":
		proto = p2.New(network)
		servo, err = xl.New(proto, *servoID)

	default:
		fmt.Printf("unsupported servo model: %s\n", *model)
	}

	if err != nil {
		fmt.Printf("servo init error: %s\n", err)
		os.Exit(1)
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
