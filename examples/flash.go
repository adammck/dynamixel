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

	network := dynamixel.NewNetwork(serial)
	network.Debug = *debug

	servo := dynamixel.NewServo(network, uint8(*servoId))
	err = servo.Ping()
	if err != nil {
		fmt.Printf("ping error: %s\n", err)
		os.Exit(1)
	}

	led := false
	for {
		led = !led
		servo.SetLed(led)
		time.Sleep(time.Duration(*interval) * time.Millisecond)
	}
}
