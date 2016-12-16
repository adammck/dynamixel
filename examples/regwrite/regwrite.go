package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/adammck/dynamixel/network"
	proto1 "github.com/adammck/dynamixel/protocol/v1"
	"github.com/adammck/dynamixel/servo/ax"
	"github.com/jacobsa/go-serial/serial"
)

var (
	portName = flag.String("port", "/dev/tty.usbserial-A9ITPZVR", "the serial port path")
	servoIDs = flag.String("id", "1,2,3", "the IDs of the servos to move (comma-separated)")
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
		network.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	network.Flush()

	for _, strID := range strings.Split(*servoIDs, ",") {
		servoID, err := strconv.Atoi(strID)
		if err != nil {
			fmt.Printf("invalid servo ID: %s (err=%s)\n", strID, err)
			os.Exit(1)
		}

		// TODO: Support XL servos, too.
		//       See examples/set-position.
		servo, err := ax.New(network, servoID)

		if err != nil {
			fmt.Printf("servo init error: %s\n", err)
			os.Exit(1)
		}

		err = servo.Ping()
		if err != nil {
			fmt.Printf("ping error: %s\n", err)
			os.Exit(1)
		}

		// Send RegWrite instead of WriteData
		servo.SetBuffered(true)

		err = servo.SetGoalPosition(*position)
		if err != nil {
			fmt.Printf("move error: %s\n", err)
			os.Exit(1)
		}
	}

	proto := proto1.New(network)
	err = proto.Action()
	if err != nil {
		fmt.Printf("action error: %s\n", err)
		os.Exit(1)
	}
}
