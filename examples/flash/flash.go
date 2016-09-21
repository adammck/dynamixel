package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

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
	servoId  = flag.Int("id", 1, "the ID of the servo to flash")
	model    = flag.String("model", "ax", "the model of the servo to flash")
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

	var nw iface.Protocol
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

	led := true
	for {
		err = servo.SetLED(led)
		if err != nil {
			fmt.Printf("SetLED error: %s\n", err)
			os.Exit(1)
		}

		time.Sleep(time.Duration(*interval) * time.Millisecond)
		led = !led
	}
}
