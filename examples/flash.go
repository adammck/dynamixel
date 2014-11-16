package main

import (
	"flag"
	"fmt"
	"github.com/adammck/dynamixel"
	"github.com/jacobsa/go-serial/serial"
	"os"
	"time"
)

var (
	portName = flag.String("port", "/dev/tty.usbserial-A9ITPZVR", "the serial port path")
	servoId  = flag.Int("id", 1, "the ID of the servo to flash")
	interval = flag.Int("interval", 200, "the time between flashes (ms)")
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
		fmt.Println(openErr)
		os.Exit(1)
	}

	network := dynamixel.NewNetwork(serial)
	servo := dynamixel.NewServo(network, uint8(*servoId))

	pingErr := servo.Ping()
	if pingErr != nil {
		fmt.Println(pingErr)
		os.Exit(1)
	}

	led := false
	for {
		led = !led
		servo.SetLed(led)
		time.Sleep(time.Duration(*interval) * time.Millisecond)
	}
}
