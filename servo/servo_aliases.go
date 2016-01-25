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
