package servo

// High-level interface. (Most of this should be removed, or moved to a separate
// type which embeds or interacts with the servo type.)

func (s *Servo) posToAngle(p int) float64 {
	return (positionToAngle * float64(p)) - s.zeroAngle
}

func (s *Servo) angleToPos(angle float64) int {
	return int((s.zeroAngle + angle) * angleToPosition)
}

// Sets the origin angle (in degrees).
func (s *Servo) SetZero(offset float64) {
	s.zeroAngle = offset
}

// Returns the current position of the servo, relative to the zero angle.
func (s *Servo) Angle() (float64, error) {
	p, err := s.Position()

	if err != nil {
		return 0, err

	} else {
		return s.posToAngle(p), nil
	}
}

// MoveTo sets the goal position of the servo by angle (in degrees), where zero
// is the midpoint, 150 deg is max left (clockwise), and -150 deg is max right
// (counter-clockwise). This is generally preferable to calling SetGoalPosition,
// which uses the internal uint16 representation.
func (s *Servo) MoveTo(angle float64) error {
	p := s.angleToPos(normalizeAngle(angle))
	return s.SetGoalPosition(p)
}

// Voltage returns the current voltage supplied. Unlike the underlying register,
// this is the actual voltage, not multiplied by ten.
func (s *Servo) Voltage() (float64, error) {
	val, err := s.PresentVoltage()
	if err != nil {
		return 0.0, err
	}

	// Convert the return value into actual volts.
	return (float64(val) / 10), nil
}

//
func normalizeAngle(d float64) float64 {
	if d > 180 {
		return normalizeAngle(d - 360)

	} else if d < -180 {
		return normalizeAngle(d + 360)

	} else {
		return d
	}
}
