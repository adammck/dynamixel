package servo

// Legacy aliases. These will eventually be removed.

// SetIdent is a legacy alias for SetServoID.
func (s *Servo) SetIdent(ident int) error {
	return s.SetServoID(ident)
}

// Position is a legacy alias for PresentPosition.
func (s *Servo) Position() (int, error) {
	return s.PresentPosition()
}

// Registered is a legacy alias for RegisteredInstruction. The XL docs use the
// longer name, which seems clearer to me.
func (s *Servo) Registered() (int, error) {
	return s.RegisteredInstruction()
}

// LimitTemperature is an alias for HighestLimitTemperature, since the XL-320
// calls it that.
func (s *Servo) LimitTemperature() (int, error) {
	return s.HighestLimitTemperature()
}

// SetLimitTemperature is an alias for SetHighestLimitTemperature, since the
// XL-320 calls it that.
func (s *Servo) SetLimitTemperature(v int) error {
	return s.SetHighestLimitTemperature(v)
}
