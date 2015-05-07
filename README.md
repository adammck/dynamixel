# Dynamixel

This packages provides a Go interface to Dynamixel servos. It's incomplete, but
the important stuff works. It's only tested against AX-12A servos (because I am
a cheapskate), but probably works with other AX models.


## Example

```go
package main

import (
  "log"
  "github.com/jacobsa/go-serial/serial"
  "github.com/adammck/dynamixel"
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

  network := dynamixel.NewNetwork(serial)
  servo := dynamixel.NewServo(network, 1)

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
The API is based on the Dynamixel [AX protocol] [proto] docs.


## License

[MIT] [license], obv.


## Author

[Adam Mckaig] [adammck] made this just for you.  




[docs]:     https://godoc.org/github.com/adammck/dynamixel
[examples]: https://github.com/adammck/dynamixel/tree/master/examples
[proto]:    http://support.robotis.com/en/product/dynamixel/ax_series/dxl_ax_actuator.htm#Control_Table
[license]:  https://github.com/adammck/dynamixel/blob/master/LICENSE
[adammck]:  http://github.com/adammck
