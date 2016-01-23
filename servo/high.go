package servo

// High-level interface. (Most of this should be removed, or moved to a separate
// type which embeds or interacts with the servo type.)

func (servo *Servo) posToAngle(pos int) float64 {
	return (positionToAngle * float64(pos)) - servo.zeroAngle
}

func (servo *Servo) angleToPos(angle float64) int {
	return int((servo.zeroAngle + angle) * angleToPosition)
}

// Sets the origin angle (in degrees).
func (servo *Servo) SetZero(offset float64) {
	servo.zeroAngle = offset
}

// Returns the current position of the servo, relative to the zero angle.
func (servo *Servo) Angle() (float64, error) {
	pos, err := servo.Position()

	if err != nil {
		return 0, err

	} else {
		return servo.posToAngle(pos), nil
	}
}

// MoveTo sets the goal position of the servo by angle (in degrees), where zero
// is the midpoint, 150 deg is max left (clockwise), and -150 deg is max right
// (counter-clockwise). This is generally preferable to calling SetGoalPosition,
// which uses the internal uint16 representation.
func (servo *Servo) MoveTo(angle float64) error {
	pos := servo.angleToPos(normalizeAngle(angle))
	return servo.SetGoalPosition(pos)
}

// Voltage returns the current voltage supplied. Unlike the underlying register,
// this is the actual voltage, not multiplied by ten.
func (servo *Servo) Voltage() (float64, error) {
	val, err := servo.PresentVoltage()
	if err != nil {
		return 0.0, err
	}

	// Convert the return value into actual volts.
	return (float64(val) / 10), nil
}
