# Dynamixel

This packages provides a Go interface to Dynamixel servos. It's brand new, but
some things kind of work.


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
    MinimumReadSize: 1,
  }

  serial, openErr := serial.Open(options); if openErr != nil {
    log.Fatalf("error opening serial port: %v\n", openErr)
  }

  network := dynamixel.NewNetwork(serial)
  servo := dynamixel.NewServo(network, 1)

  moveErr := servo.SetGoalPosition(512); if moveErr != nil {
    log.Fatalf("error setting goal position: %v\n", moveErr)
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
