package servo

// Legacy aliases. These will eventually be removed.

// SetIdent is a legacy alias for SetServoID.
func (servo *Servo) SetIdent(ident int) error {
	return servo.SetServoID(ident)
}

// Position is a legacy alias for PresentPosition.
func (servo *Servo) Position() (int, error) {
	return servo.PresentPosition()
}
