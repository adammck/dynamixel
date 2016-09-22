# Dynamixel

[![Build Status](https://travis-ci.org/adammck/dynamixel.svg?branch=master)](https://travis-ci.org/adammck/dynamixel)
[![Go Report Card](https://goreportcard.com/badge/github.com/adammck/dynamixel)](https://goreportcard.com/report/github.com/adammck/dynamixel)
[![GoDoc](https://godoc.org/github.com/adammck/dynamixel?status.svg)](https://godoc.org/github.com/adammck/dynamixel)

This packages provides a Go interface to Dynamixel servos. It's tested against
[AX-12] [ax] and [XL-320] [xl] servos (because I am a cheapskate), but should
work with other models.


## Example

```go
package main

import (
  "log"
  "github.com/jacobsa/go-serial/serial"
  "github.com/adammck/dynamixel/network"
  "github.com/adammck/dynamixel/servo/ax"
)

func main() {
  options := serial.OpenOptions{
    PortName: "/dev/tty.usbserial-A9ITPZVR",
    BaudRate: 1000000,
    DataBits: 8,
    StopBits: 1,
    MinimumReadSize: 0,
    InterCharacterTimeout: 100,
  }

  serial, err := serial.Open(options)
  if err != nil {
    log.Fatalf("error opening serial port: %v\n", err)
  }

  network := network.New(serial)
  servo, err := ax.New(network, 1)
  if err != nil {
    log.Fatalf("error initializing servo: %v\n", err)
  }

  err = servo.Ping()
  if err != nil {
    log.Fatalf("error pinging servo: %v\n", err)
  }

  err = servo.SetGoalPosition(512)
  if err != nil {
    log.Fatalf("error setting goal position: %v\n", err)
  }
}
```

More examples can be found in the [examples] [examples] directory of this repo.


## Documentation

The docs can be found at [godoc.org] [docs], as usual.  
The API is based on the Dynamixel [v1 protocol] [proto] docs.


## License

[MIT] [license], obv.


## Author

[Adam Mckaig] [adammck] made this just for you.  




[ax]:       http://support.robotis.com/en/product/dynamixel/ax_series/dxl_ax_actuator.htm
[xl]:       http://support.robotis.com/en/product/dynamixel/xl-series/xl-320.htm
[docs]:     https://godoc.org/github.com/adammck/dynamixel
[examples]: https://github.com/adammck/dynamixel/tree/master/examples
[proto]:    http://support.robotis.com/en/product/dynamixel/ax_series/dxl_ax_actuator.htm#Control_Table
[license]:  https://github.com/adammck/dynamixel/blob/master/LICENSE
[adammck]:  http://github.com/adammck
