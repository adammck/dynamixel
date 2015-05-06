package main

import (
	"flag"
	"fmt"
	"github.com/adammck/dynamixel"
	"github.com/jacobsa/go-serial/serial"
	"os"
)

var (
	portName = flag.String("port", "/dev/tty.usbserial-A9ITPZVR", "the serial port path")
	servoId  = flag.Int("id", 1, "the ID of the servo to move")
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

	serial, openErr := serial.Open(options)
	if openErr != nil {
		fmt.Printf("open error: %s\n", openErr)
		os.Exit(1)
	}

	network := dynamixel.NewNetwork(serial)
	network.Debug = *debug

	servo := dynamixel.NewServo(network, uint8(*servoId))
	pingErr := servo.Ping()
	if pingErr != nil {
		fmt.Println("ping error: %s\n", pingErr)
		os.Exit(1)
	}

	moveErr := servo.SetGoalPosition(*position)
	if moveErr != nil {
		fmt.Println("move error: %s\n", moveErr)
		os.Exit(1)
	}
}
