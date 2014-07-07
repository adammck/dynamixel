# Dynamixel

This packages provides a Go interface to Dynamixel servos. It's brand new and
nothing works yet.


## Example

```go
package main

import (
  "log"
  "time"
  "math/rand"
  "github.com/adammck/serial"
  "github.com/adammck/dynamixel"
)

func main() {

  options := serial.OpenOptions{
    PortName: "/dev/tty.usbserial-A9ITPZVR",
    BaudRate: 1000000,
    DataBits: 8,
    StopBits: 1,
    MinimumReadSize: 4,
  }

  serial, err := serial.Open(options)
  if err != nil {
    log.Fatal(err)
  }

  network := dynamixel.NewNetwork(serial)
  servo := dynamixel.NewServo(network, 1)

  for {
    // This doesn't actually work yet!
    servo.SetGoalPosition(rand.Intn(1024))
    time.Sleep(2 * time.Second)
  }
}
```


## License

[MIT] (https://github.com/adammck/dynamixel/blob/master/LICENSE), obv.


## Author

[Adam Mckaig] (http://github.com/adammck) made this just for you.
