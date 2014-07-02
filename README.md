# Dynamixel

This packages provides a Go interface to Dynamixel servos. It's brand new and
nothing works yet.


## Example

```go
package main

import (
  "time"
  "math/rand"
  "github.com/adammck/dynamixel"
)

func main() {

  // This doesn't actually work yet!
  network := dynamixel.NewNetwork("/dev/cu.usbserial-A9ITPZVR")
  servo := dynamixel.NewServo(network, 1)

  for {
    // Nor does this!
    servo.SetGoalPosition(rand.Intn(1024))
    time.Sleep(2 * time.Second)
  }
}
```


## License

[MIT] (https://github.com/adammck/dynamixel/blob/master/LICENSE), obv.


## Author

[Adam Mckaig] (http://github.com/adammck) made this just for you.
