package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/adammck/dynamixel/network"
	"github.com/adammck/dynamixel/servo/xl"
	"github.com/jacobsa/go-serial/serial"
)

var (
	portName = flag.String("port", "/dev/tty.usbserial-A9ITPZVR", "the serial port path")
	oldIdent = flag.Int("old", 1, "the current ID of the servo")
	newIdent = flag.Int("new", 1, "the new ID to write")
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
		network.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	network.Flush()

	servo, err := xl.New(network, *oldIdent)
	if err != nil {
		fmt.Printf("servo init error: %s\n", err)
		os.Exit(1)
	}

	err = servo.Ping()
	if err != nil {
		fmt.Printf("ping error: %s\n", err)
		os.Exit(1)
	}

	torque, err := servo.TorqueEnable()
	if err != nil {
		fmt.Printf("torque enabled error: %s\n", err)
		os.Exit(1)
	}

	if torque {
		err = servo.SetTorqueEnable(false)
		if err != nil {
			fmt.Printf("set torque enabled error: %s\n", err)
			os.Exit(1)
		}
	}

	err = servo.SetServoID(*newIdent)
	if err != nil {
		fmt.Printf("set return delay time error: %s\n", err)
		os.Exit(1)
	}

	if torque {
		err = servo.SetTorqueEnable(true)
		if err != nil {
			fmt.Printf("set torque enabled error: %s\n", err)
			os.Exit(1)
		}
	}
}
